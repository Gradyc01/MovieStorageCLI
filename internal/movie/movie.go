package movie

import (
	"fmt"
	"movie-tracker/internal/api"
	"strconv"
	"strings"
	"time"
	"unicode"

	"regexp"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9-]`)

// Movie is our domain object. Capitalized field names are "exported"
// (public) — visible to any package that imports this one. If a field
// started with a lowercase letter, it would be private to this package,
// similar to Java's default/package-private access.
//
// The `json:"..."` text after each field is a "struct tag" — metadata
// read by the encoding/json package (and others) to know how to name
// this field when converting to/from JSON. Think of it like Jackson's
// @JsonProperty annotation in Java, but built into the language.
type Movie struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Year        int       `json:"year"`
	ReleaseDate string    `json:"release_date"` //Has default value of ""
	Watched     bool      `json:"watched"`      //Has default value of false
	Rating      float64   `json:"rating"`       //Has default value of -1
	Franchise   string    `json:"franchise"`    //Has default value of ""
	Directors   []string  `json:"director"`     //Has default value of ""
	ImdbID      string    `json:"imdb_id"`      //Has default value of ""
	AddedAt     time.Time `json:"added_at"`
}

// NewMovie is our "constructor". Go doesn't have constructors as a
// language feature — this is just a regular function that returns a
// ready-to-use Movie. Returning a pointer (*Movie) means callers get
// a reference to one shared instance, similar to how Java objects
// are always references under the hood.
func NewMovie(title string, releaseDate string, franchise string, directors []string, imdbID string, rating float64) *Movie {
	year := getReleaseYear(releaseDate)
	watched := rating != -1
	return &Movie{
		ID:          generateID(title, year),
		Title:       title,
		Year:        year,
		ReleaseDate: releaseDate,
		Watched:     watched,
		Rating:      rating,
		Franchise:   franchise,
		Directors:   directors,
		ImdbID:      imdbID,
		AddedAt:     time.Now(),
	}
}

func NewMovieViaImdbLink(imdbLink string, rating float64) (*Movie, error) {
	client := api.NewClientUsingEnvVariable()

	imdbID, err := api.ExtractIMDbID(imdbLink)
	if err != nil {
		return nil, err
	}
	result, err := client.FindByIMDbID(imdbID)
	if err != nil {
		return nil, fmt.Errorf("error getting movie by IMDB URL: %w", err)
	}

	if len(result.MovieResults) > 0 {
		movie, err := client.GetMovie(result.MovieResults[0].ID)
		if err != nil {
			return nil, fmt.Errorf("error getting movie by IMDB URL: %w", err)
		}
		return NewMovie(movie.Title, movie.ReleaseDate, findFranchise(movie.BelongsToCollection), findDirector(client, result.MovieResults[0].ID), imdbID, rating), nil
	}
	return nil, fmt.Errorf("no match found on TMDB for that IMDb ID")
}

func NewShowViaImdbLink(imdbLink string, seasonNumber int, rating float64) (*Movie, error) {
	client := api.NewClientUsingEnvVariable()

	imdbID, err := api.ExtractIMDbID(imdbLink)
	if err != nil {
		return nil, err
	}
	result, err := client.FindByIMDbID(imdbID)
	if err != nil {
		return nil, fmt.Errorf("error getting movie by IMDB URL: %w", err)
	}

	if len(result.TVResults) > 0 {
		tv, err := client.GetTVShow(result.TVResults[0].ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tv by IMDB URL: %w", err)
		}

		season, err := findSeason(tv, seasonNumber)
		if err != nil {
			return nil, err
		}

		return NewMovie(tv.Name+" "+season.Name, season.AirDate, findFranchise(tv.BelongsToCollection), findCreators(tv.CreatedBy), imdbID, rating), nil
	}
	return nil, fmt.Errorf("no match found on TMDB for that IMDb ID")
}

func generateID(title string, year int) string {
	// Normalize to NFD to separate base characters from diacritics (e.g., ō -> o + combining mark)
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	cleanedTitle, _, _ := transform.String(t, title)

	// Convert to lowercase, remove spaces, and strip all characters except a-z, 0-9, and '-'
	slug := strings.ToLower(cleanedTitle)
	slug = strings.ReplaceAll(slug, " ", "")
	slug = slugRegexp.ReplaceAllString(slug, "")

	return fmt.Sprintf("%s-%d", slug, year)
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Identifies non-spacing diacritical marks
}

// Obtains release year through the release date.
// String releaseDate must follow format YYYY-MM-DD
func getReleaseYear(releaseDate string) int {
	year, err := strconv.ParseInt(strings.Split(releaseDate, "-")[0], 10, 64)
	if err != nil {
		fmt.Printf("error parsing release year: %v\n", err)
		return -1
	}
	return int(year)
}

func findFranchise(collection *api.Collection) string {
	if collection != nil {
		return collection.Name
	}
	return ""
}

func findDirector(client *api.Client, tmdbID int) []string {
	result, err := client.GetMovieCredits(tmdbID)
	if err != nil {
		fmt.Printf("error getting movie credits: %v\n", err)
	}

	var directors []string
	for _, crew := range result.Crew {
		if crew.Job == "Director" {
			directors = append(directors, crew.Name)
		}
	}
	return directors
}

func findCreators(people []api.Person) []string {
	var creators []string
	for _, person := range people {
		creators = append(creators, person.Name)
	}
	return creators
}

func findSeason(show *api.TVShow, seasonNumber int) (api.Season, error) {
	for _, season := range show.Seasons {
		if season.SeasonNumber == seasonNumber {
			return season, nil
		}
	}
	return api.Season{}, fmt.Errorf("no season found with season number %d", seasonNumber)
}

// String implements the fmt.Stringer interface. In Go, if a type has
// a method called String() string, then fmt.Println(m) and similar
// will automatically use it — no need to declare "implements Stringer"
// anywhere. This is "structural typing": you satisfy an interface just
// by having the right method signature.
//
// Notice the receiver: (m *Movie). This means String() is a method
// attached to *Movie, roughly like an instance method in Java, where
// `m` plays the role of `this`.
func (m *Movie) String() string {
	status := "unwatched"
	if m.Watched {
		status = "watched"
	}
	return fmt.Sprintf("%s (%d) - %s", m.Title, m.Year, status)
}

// MarkWatched is another method on *Movie. Methods that mutate the
// receiver need a pointer receiver (*Movie), not a value receiver
// (Movie), or the change won't stick — similar in spirit to needing
// a reference (not a copy) in Java to mutate shared state.
func (m *Movie) MarkWatched() {
	m.Watched = true
}

// DirectorsDisplay joins the Directors slice into a single
// comma-separated string, or a placeholder if none are set. Small
// formatting helpers like this are common in Go — since there's no
// method overloading, it's normal to add a purpose-named method
// rather than trying to cram every case into String().
func (m *Movie) DirectorsDisplay() string {
	if len(m.Directors) == 0 {
		return "—"
	}
	return strings.Join(m.Directors, ", ")
}

// RatingOrWatched implements the combined column your list view wants:
// if the movie hasn't been watched yet, show "Unwatched"; otherwise
// show the numeric rating (or "Unrated" if watched but never scored).
func (m *Movie) RatingOrWatched() string {
	if !m.Watched {
		return "Unwatched"
	}
	if m.Rating < 0 {
		return "Unrated"
	}
	return fmt.Sprintf("%.1f/10", m.Rating)
}
