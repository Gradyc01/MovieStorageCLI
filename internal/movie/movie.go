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
	Version             int       `json:"version"`
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	Year                int       `json:"year"`
	ReleaseDate         string    `json:"release_date"` //Has default value of ""
	Watched             bool      `json:"watched"`      //Has default value of false
	Rating              float64   `json:"rating"`       //Has default value of -1
	Directors           []string  `json:"director"`     //Has default value of ""
	ImdbID              string    `json:"imdb_id"`      //Has default value of ""
	AddedAt             time.Time `json:"added_at"`
	Status              string    `json:"status"` //Can be of value watched, unwatched, unrated, shortlist
	Genres              []string  `json:"genres"`
	FilmType            string    `json:"film_type"` //Can be of value TV or Movie
	Tags                []string  `json:"tags"`
	ProductionCompanies []string  `json:"production_companies"`
	ProductionCountries []string  `json:"production_countries"`
	SpokenLanguages     []string  `json:"spoken_languages"`
	KnownActors         []string  `json:"known_actors"`
	TmdbID              string    `json:"tmdb_id"`
	Notes               string    `json:"notes"`
}

// NewMovie is our "constructor". Go doesn't have constructors as a
// language feature — this is just a regular function that returns a
// ready-to-use Movie. Returning a pointer (*Movie) means callers get
// a reference to one shared instance, similar to how Java objects
// are always references under the hood.
func NewMovie(title string, releaseDate string, directors []string, imdbID string, tmdbID string, genres []string, filmType string, rating float64, tags []string,
	prodCompanies []string, prodCountries []string, languages []string, actors []string, notes string) *Movie {
	year := getReleaseYear(releaseDate)
	watched := rating != -1
	status := "unwatched"
	if rating != -1 {
		status = "watched"
	}
	return &Movie{
		Version:             1,
		ID:                  generateID(title, year),
		Title:               title,
		Year:                year,
		ReleaseDate:         releaseDate,
		Watched:             watched,
		Rating:              rating,
		Directors:           directors,
		ImdbID:              imdbID,
		AddedAt:             time.Now(),
		Status:              status,
		Genres:              genres,
		FilmType:            filmType,
		Tags:                tags,
		ProductionCompanies: prodCompanies,
		ProductionCountries: prodCountries,
		SpokenLanguages:     languages,
		KnownActors:         actors,
		TmdbID:              tmdbID,
		Notes:               notes,
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

	tmdbID := result.MovieResults[0].ID

	credits, err := client.GetMovieCredits(tmdbID)
	if err != nil {
		fmt.Printf("error getting movie credits: %v\n", err)
	}

	if len(result.MovieResults) > 0 {
		movie, err := client.GetMovie(tmdbID)
		if err != nil {
			return nil, fmt.Errorf("error getting movie by IMDB URL: %w", err)
		}
		return NewMovie(
			movie.Title,
			movie.ReleaseDate,
			findDirector(credits),
			imdbID,
			strconv.Itoa(tmdbID),
			findGenres(movie.Genres),
			"movie",
			rating,
			findFranchise(movie.BelongsToCollection),
			findProductionCompanies(movie.ProductionCompanies),
			findProductionCountries(movie.ProductionCountries),
			findSpokenLanguages(movie.SpokenLanguages),
			findKnownActors(credits),
			"",
		), nil
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
		return nil, fmt.Errorf("error getting tv by IMDB URL: %w", err)
	}

	tmdbID := result.TVResults[0].ID

	credits, err := client.GetTVCredits(tmdbID)
	if err != nil {
		fmt.Printf("error getting tv credits: %v\n", err)
	}

	if len(result.TVResults) > 0 {
		tv, err := client.GetTVShow(tmdbID)
		if err != nil {
			return nil, fmt.Errorf("error getting tv by IMDB URL: %w", err)
		}

		season, err := findSeason(tv, seasonNumber)
		if err != nil {
			return nil, err
		}

		return NewMovie(
				tv.Name+" "+season.Name,
				season.AirDate,
				findCreators(tv.CreatedBy),
				imdbID,
				strconv.Itoa(tmdbID),
				findGenres(tv.Genres),
				"tv",
				rating,
				findFranchise(tv.BelongsToCollection),
				findProductionCompanies(tv.ProductionCompanies),
				findProductionCountries(tv.ProductionCountries),
				findSpokenLanguages(tv.SpokenLanguages),
				findKnownActors(credits),
				""),
			nil
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

func findProductionCompanies(companies []api.ProductionCompany) []string {
	var companiesList []string
	for _, company := range companies {
		companiesList = append(companiesList, company.Name)
	}
	return companiesList
}

func findProductionCountries(countries []api.ProductionCountry) []string {
	var countryList []string
	for _, country := range countries {
		countryList = append(countryList, country.Name)
	}
	return countryList
}

func findSpokenLanguages(languages []api.SpokenLanguage) []string {
	var languageList []string
	for _, language := range languages {
		languageList = append(languageList, language.EnglishName)
	}
	return languageList
}

func findFranchise(collection *api.Collection) []string {
	var franchises []string
	if collection != nil {
		franchises = append(franchises, collection.Name)
	}
	return franchises
}

func findGenres(genres []api.Genres) []string {
	var genreList []string
	for _, genre := range genres {
		genreList = append(genreList, genre.Name)
	}
	return genreList
}

func findDirector(credits *api.Credits) []string {
	var directors []string
	for _, crew := range credits.Crew {
		if crew.Job == "Director" {
			directors = append(directors, crew.Name)
		}
	}
	return directors
}

func findKnownActors(credits *api.Credits) []string {
	var actors []string
	for _, actor := range credits.Cast {
		if actor.Popularity > 1.0 {
			actors = append(actors, actor.Name)
		}
	}
	return actors
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

// ListDisplay joins the Directors slice into a single
// comma-separated string, or a placeholder if none are set. Small
// formatting helpers like this are common in Go — since there's no
// method overloading, it's normal to add a purpose-named method
// rather than trying to cram every case into String().
func (m *Movie) ListDisplay(list []string) string {
	if len(list) == 0 {
		return "—"
	}
	return strings.Join(list, ", ")
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
