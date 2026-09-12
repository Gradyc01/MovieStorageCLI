package github

import (
	"context"
	"fmt"
	"movie-tracker/internal/environment"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Store wraps the GitHub client + the bits of state you need to safely
// read-modify-write the file (owner/repo/path + the last-seen SHA).
type Store struct {
	client    *github.Client
	owner     string
	repo      string
	path      string // path to the file *within the repo*
	localPath string // path to the local cached copy on disk
	branch    string // usually "main"
	sha       string // set after every Load()/Save(), required for the next Save()
}

// LocalPath returns where the cached JSON file lives on disk, so the
// rest of your app (which already knows how to read/write that file)
// doesn't need to know anything about GitHub.
func (s *Store) LocalPath() string {
	return s.localPath
}

func NewStore() (*Store, error) {
	token, err1 := environment.GetVariable("GITHUB_TOKEN")
	repoURL, err2 := environment.GetVariable("MOVIE_REPO") // e.g. "yourname/movie-list"
	path, err3 := environment.GetVariable("MOVIE_FILE_PATH")
	if err1 != nil {
		path = "movies.json"
	}
	if err2 != nil || err3 != nil {
		return nil, fmt.Errorf("GITHUB_TOKEN and MOVIE_REPO must be set")
	}

	owner, repo, err := splitOwnerRepo(repoURL) // "owner/repo" -> ("owner", "repo")
	if err != nil {
		return nil, err
	}

	localPath := os.Getenv("MOVIE_LOCAL_PATH")
	if localPath == "" {
		localPath = "movies.json" // same file your app already reads/writes locally
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), ts)

	return &Store{
		client:    github.NewClient(httpClient),
		owner:     owner,
		repo:      repo,
		path:      path,
		localPath: localPath,
		branch:    "main",
	}, nil
}

// Load fetches the current file from GitHub, decodes it, and stashes
// the SHA so Save() knows what version it's updating.
func (s *Store) Load(ctx context.Context) error {

	fileContent, _, resp, err := s.client.Repositories.GetContents(
		ctx, s.owner, s.repo, s.path,
		&github.RepositoryContentGetOptions{Ref: s.branch},
	)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// File doesn't exist yet — first run. No SHA yet; Save() will
			// create the file instead of updating it. Leave local cache
			// alone (or write an empty JSON array, your call).
			s.sha = ""
			return os.WriteFile(s.localPath, []byte("[]"), 0644)
		}
		return fmt.Errorf("fetching movie list: %w", err)
	}

	raw, err := fileContent.GetContent() // go-github handles the base64 decode
	if err != nil {
		return err
	}

	s.sha = fileContent.GetSHA()
	fmt.Printf("SHA: %s\n", s.sha)

	if err := os.WriteFile(s.localPath, []byte(raw), 0644); err != nil {
		return fmt.Errorf("writing local cache file: %w", err)
	}
	return nil
}

// Save pushes the updated list back as a new commit. Uses the SHA from
// the last Load() so GitHub rejects it (409) if something else changed
// the file in between — protects against silently clobbering data.
func (s *Store) Save(ctx context.Context) error {
	data, err := os.ReadFile(s.localPath)
	if err != nil {
		return fmt.Errorf("reading local cache file: %w", err)
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("update movie list"),
		Content: data,
		Branch:  github.String(s.branch),
	}
	if s.sha != "" {
		opts.SHA = github.String(s.sha) // updating an existing file
	}
	// if s.sha == "", CreateFile creates it fresh (first run case)

	resp, _, err := s.client.Repositories.UpdateFile(ctx, s.owner, s.repo, s.path, opts)
	if err != nil {
		// TODO: on 409 conflict, re-Load(), which overwrites your local
		// file — you'd lose the local changes unless you diff/reapply
		// them first. Worth thinking about if multi-device use is a
		// realistic scenario for you.
		return fmt.Errorf("saving movie list: %w", err)
	}

	s.sha = resp.GetSHA() // update for any subsequent save in the same run
	return nil
}

// //func main() {
// //	ctx := context.Background()
// //
// //	store, err := NewStore()
// //	if err != nil {
// //		fmt.Fprintln(os.Stderr, "setup error:", err)
// //		os.Exit(1)
// //	}
// //
// //	// 1. Pull latest state on open
// //	movies, err := store.Load(ctx)
// //	if err != nil {
// //		fmt.Fprintln(os.Stderr, "load error:", err)
// //		os.Exit(1)
// //	}
// //
// //	// 2. Run your existing CLI logic against `movies` in memory.
// //	//    (add/remove/query commands mutate this slice exactly like
// //	//    your current local-JSON version does — no change needed there.)
// //	//    Since add/remove may reassign the slice header (append, or
// //	//    filtering out an element), runCLI takes a pointer to the slice
// //	//    so those mutations are visible here for the Save() call.
// //	runCLI(&movies) // TODO: your existing command loop / cobra commands / etc.
// //
// //	// 3. Push changes back on close
// //	if err := store.Save(ctx, movies); err != nil {
// //		fmt.Fprintln(os.Stderr, "save error:", err)
// //		os.Exit(1)
// //	}
// //}
//
// splitOwnerRepo accepts either a bare "owner/repo" string or a full
// GitHub URL (https://github.com/owner/repo, with or without .git,
// trailing slash, etc.) and returns the owner and repo name.
func splitOwnerRepo(s string) (owner, repo string, err error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")

	//Strip protocol + host if a full URL was given.
	if idx := strings.Index(s, "github.com/"); idx != -1 {
		s = s[idx+len("github.com/"):]
	}
	s = strings.TrimPrefix(s, "git@github.com:") // covers SSH-style URLs too

	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo reference %q, expected \"owner/repo\" or a github.com URL", s)
	}
	return parts[0], parts[1], nil
}
