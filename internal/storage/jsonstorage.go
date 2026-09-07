// json_storage.go is our first concrete Storage implementation. It
// keeps every movie in a single JSON file on disk. There's no database
// here at all — we just read the whole file into memory, mutate it,
// and write the whole thing back out. That's fine for a personal CLI
// with a few hundred movies; it would NOT scale to a real production
// system, but it's a great way to learn Go's file I/O and JSON handling
// before adding SQLite complexity.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"movie-tracker/internal/movie"
)

// JSONStorage is our concrete type. In Java terms, this is like
// "class JSONStorage implements Storage" — except we never write those
// words anywhere. As long as JSONStorage has the four methods the
// Storage interface demands (Add, Remove, List, Search), the Go
// compiler considers it a valid Storage wherever one is expected.
type JSONStorage struct {
	filePath string
}

// NewJSONStorage is the constructor. Convention: New<TypeName>.
func NewJSONStorage(filePath string) *JSONStorage {
	return &JSONStorage{filePath: filePath}
}

// load reads the JSON file into a slice of *movie.Movie. It's
// unexported (lowercase) because it's an internal helper — callers of
// this package should only ever go through the Storage interface
// methods below, not reach in and call load()/save() directly.
func (s *JSONStorage) load() ([]*movie.Movie, error) {
	// os.ReadFile reads the whole file into a []byte. This is Go's
	// equivalent of Files.readAllBytes() in Java.
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// os.IsNotExist checks the specific error type. If the file
		// simply doesn't exist yet (first run), that's not really an
		// error for us — treat it as "no movies yet" and return an
		// empty slice instead of propagating the error.
		if os.IsNotExist(err) {
			return []*movie.Movie{}, nil
		}
		// Wrap the error with context using %w. This is Go's version
		// of Java's "throw new RuntimeException("reading file", e)" —
		// it preserves the original error so callers can still inspect
		// it (via errors.Is/errors.As) while adding a human-readable
		// trail of what was happening when it failed.
		return nil, fmt.Errorf("reading storage file: %w", err)
	}

	// An empty file isn't valid JSON, but it IS a valid "no movies yet"
	// state for us, so handle it explicitly rather than letting
	// json.Unmarshal fail on it.
	if len(data) == 0 {
		return []*movie.Movie{}, nil
	}

	var movies []*movie.Movie
	// json.Unmarshal parses the []byte into the given Go value. The
	// second argument (&movies) is a pointer because Unmarshal needs
	// to write INTO your variable — same reason Java's ObjectMapper
	// needs a Class<T> reference for reflection-based hydration, just
	// achieved differently here (via a pointer instead of reflection
	// on a class token).
	if err := json.Unmarshal(data, &movies); err != nil {
		return nil, fmt.Errorf("parsing storage file: %w", err)
	}
	return movies, nil
}

// save writes the given slice back to the JSON file, overwriting
// whatever was there before. This "read everything, mutate, write
// everything" pattern is simple but not safe for concurrent access —
// fine for a single-user CLI, not fine for a multi-process system.
func (s *JSONStorage) save(movies []*movie.Movie) error {
	// MarshalIndent produces pretty-printed JSON (2-space indent) so
	// the file is human-readable if you ever open it directly — useful
	// while learning/debugging.
	data, err := json.MarshalIndent(movies, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding movies: %w", err)
	}

	// 0644 is a Unix file permission (owner read/write, everyone else
	// read-only). If you've never seen this before: it's the same
	// permission model `chmod` uses.
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing storage file: %w", err)
	}
	return nil
}

// Add appends a new movie and persists it. Note the receiver type:
// (s *JSONStorage) — pointer receiver, consistent with save()/load()
// needing to eventually support mutation. As a rule of thumb: if any
// method on a type needs a pointer receiver, make them all pointer
// receivers for consistency.
func (s *JSONStorage) Add(m *movie.Movie) error {
	movies, err := s.load()
	if err != nil {
		return err
	}

	// Guard against duplicate IDs. This loop is Go's version of a
	// for-each over a List<Movie> in Java — `range` gives you
	// (index, value) pairs; we ignore the index with `_`.
	for _, existing := range movies {
		if existing.ID == m.ID {
			return fmt.Errorf("movie with id %q already exists", m.ID)
		}
	}

	movies = append(movies, m)
	return s.save(movies)
}

// Remove deletes a movie by ID. Go has no built-in "remove element
// from slice" function — you rebuild a new slice excluding the one
// you don't want. It looks unusual coming from Java's list.remove(),
// but it's the idiomatic pattern here.
func (s *JSONStorage) Remove(id string) error {
	movies, err := s.load()
	if err != nil {
		return err
	}

	found := false
	remaining := make([]*movie.Movie, 0, len(movies))
	for _, m := range movies {
		if m.ID == id {
			found = true
			continue // skip adding this one to `remaining`
		}
		remaining = append(remaining, m)
	}

	if !found {
		return fmt.Errorf("no movie found with id %q", id)
	}

	return s.save(remaining)
}

// List returns every stored movie, no filtering.
func (s *JSONStorage) List() ([]*movie.Movie, error) {
	return s.load()
}

// Search does a simple case-insensitive substring match on the title.
// Nothing fancy — this is a placeholder you can upgrade later (fuzzy
// matching, searching by director/franchise too, etc.).
func (s *JSONStorage) Search(query string) ([]*movie.Movie, error) {
	movies, err := s.load()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*movie.Movie
	for _, m := range movies {
		if strings.Contains(strings.ToLower(m.Title), query) {
			results = append(results, m)
		}
	}
	return results, nil
}

// GetByID returns the single movie matching id, or an error if none
// is found. Straightforward linear scan — fine at this data size;
// if this ever became a bottleneck, an index (map[string]*movie.Movie)
// built once at load time would be the next step, but that's not
// worth the added complexity yet.
func (s *JSONStorage) GetByID(id string) (*movie.Movie, error) {
	movies, err := s.load()
	if err != nil {
		return nil, err
	}

	for _, m := range movies {
		if m.ID == id {
			return m, nil
		}
	}

	return nil, fmt.Errorf("no movie found with id %q", id)
}
