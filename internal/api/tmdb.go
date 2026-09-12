// Package main provides a Go client for The Movie Database (TMDB) API,
// focused on looking up a Movie or TV Show from a pasted IMDb URL/ID via
// TMDB's "find by external ID" endpoint.
//
// Setup:
//  1. Get a free API key (v3 auth) or read access token (v4 auth) from:
//     https://www.themoviedb.org/settings/api
//  2. Create environment.properties in the project root with:
//     TMDB_API_KEY=your_api_key_here
//  3. Run: go run tmdb_client.go
//
// This example uses v3 "api_key" query-param auth for simplicity. If you'd
// rather use a v4 Bearer token, see the NewClientWithBearerToken constructor
// below.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"movie-tracker/internal/environment"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL         = "https://api.themoviedb.org/3"
	imageBaseURL    = "https://image.tmdb.org/t/p"
	BearerTokenName = "TMDB_API_KEY"
)

// Client is a small wrapper around TMDB's REST API.
type Client struct {
	apiKey      string // v3 auth (query param)
	bearerToken string // v4 auth (Authorization header)
	httpClient  *http.Client
}

// NewClient creates a client using v3 "api_key" query-param authentication.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func NewClientUsingEnvVariable() *Client {
	key, err := environment.GetVariable(BearerTokenName)
	if err != nil {
		fmt.Printf("Error getting API key: %s\n", err)
		os.Exit(1)
		return nil
	}
	return NewClientWithBearerToken(key)
}

// NewClientWithBearerToken creates a client using v4 Bearer token authentication.
func NewClientWithBearerToken(token string) *Client {
	return &Client{
		bearerToken: token,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// get performs a GET request against the TMDB API and unmarshals the JSON
// response into `out`. `params` is a map of extra query-string parameters.
func (c *Client) get(path string, params map[string]string, out interface{}) error {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	if c.apiKey != "" {
		q.Set("api_key", c.apiKey)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			StatusMessage string `json:"status_message"`
			StatusCode    int    `json:"status_code"`
		}
		_ = json.Unmarshal(body, &apiErr)
		return fmt.Errorf("tmdb api error (http %d): %s", resp.StatusCode, apiErr.StatusMessage)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// ---------- Data models ----------

// Collection represents the "belongs_to_collection" object returned for some movies & shows
type Collection struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// Movie represents a subset of fields returned for movie results/details.
type Movie struct {
	ID                  int                 `json:"id"`
	Title               string              `json:"title"`
	OriginalTitle       string              `json:"original_title"`
	Overview            string              `json:"overview"`
	ReleaseDate         string              `json:"release_date"`
	PosterPath          string              `json:"poster_path"`
	BackdropPath        string              `json:"backdrop_path"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`
	Popularity          float64             `json:"popularity"`
	Adult               bool                `json:"adult"`
	GenreIDs            []int               `json:"genre_ids,omitempty"`
	Runtime             int                 `json:"runtime,omitempty"`
	Tagline             string              `json:"tagline,omitempty"`
	Status              string              `json:"status,omitempty"`
	OriginalLanguage    string              `json:"original_language"`
	BelongsToCollection *Collection         `json:"belongs_to_collection"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Genres              []Genres            `json:"genres"`
}

// TVShow represents a subset of fields returned for TV results/details.
type TVShow struct {
	ID                  int                 `json:"id"`
	Name                string              `json:"name"`
	OriginalName        string              `json:"original_name"`
	Overview            string              `json:"overview"`
	FirstAirDate        string              `json:"first_air_date"`
	PosterPath          string              `json:"poster_path"`
	BackdropPath        string              `json:"backdrop_path"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`
	Popularity          float64             `json:"popularity"`
	OriginalLanguage    string              `json:"original_language"`
	BelongsToCollection *Collection         `json:"belongs_to_collection"`
	CreatedBy           []Person            `json:"created_by"`
	Seasons             []Season            `json:"seasons"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Genres              []Genres            `json:"genres"`
}

// Credits represents a subset of fields returned for credits of TV or Movie.
type Credits struct {
	ID   int      `json:"id"`
	Cast []Person `json:"cast"`
	Crew []Crew   `json:"crew"`
}

// ProductionCompany represents a struct representing a production company
type ProductionCompany struct {
	ID       int    `json:"id"`
	LogoPath string `json:"logo_path"`
	Name     string `json:"name"`
	Country  string `json:"origin_country"`
}

// ProductionCountry represents a struct representing a production country
type ProductionCountry struct {
	ISO31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

// SpokenLanguage represents a struct representing a spoken language
type SpokenLanguage struct {
	EnglishName string `json:"english_name"`
	Name        string `json:"name"`
	ISO6391     string `json:"iso_639_1"`
}

// Genres represents a struct representing a Genre
type Genres struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Person represents a subset of fields returned for people results/details.
type Person struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`
}

// Crew represents a subset of fields returned for crew results/details.
type Crew struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`
	Department  string  `json:"department"`
	Job         string  `json:"job"`
}

// PagedMovies is TMDB's paginated response envelope for movie lists.
type PagedMovies struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

// PagedTVShows is TMDB's paginated response envelope for TV lists.
type PagedTVShows struct {
	Page         int      `json:"page"`
	Results      []TVShow `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

// PagedPeople is TMDB's paginated response envelope for people lists.
type PagedPeople struct {
	Page         int      `json:"page"`
	Results      []Person `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

// ---------- Find-by-external-ID endpoint ----------

// FindResults is TMDB's response envelope for the /find endpoint. TMDB
// doesn't know in advance whether an IMDb ID refers to a movie, a TV show,
// or an episode, so it returns all matching buckets and only the relevant
// one will be populated.
type FindResults struct {
	MovieResults     []Movie   `json:"movie_results"`
	TVResults        []TVShow  `json:"tv_results"`
	TVEpisodeResults []Episode `json:"tv_episode_results"`
	TVSeasonResults  []Season  `json:"tv_season_results"`
	PersonResults    []Person  `json:"person_results"`
}

// Episode represents a subset of fields for a TV episode find-result.
type Episode struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	AirDate       string  `json:"air_date"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	ShowID        int     `json:"show_id"`
	VoteAverage   float64 `json:"vote_average"`
}

// Season represents a subset of fields for a TV season find-result.
type Season struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	AirDate      string `json:"air_date"`
	SeasonNumber int    `json:"season_number"`
	ShowID       int    `json:"show_id"`
}

// imdbIDPattern matches an IMDb title/episode ID like "tt1375666" anywhere
// in a string (e.g. embedded in a full IMDb URL).
var imdbIDPattern = regexp.MustCompile(`tt\d{6,10}`)

// ExtractIMDbID pulls the "tt#######" ID out of a full IMDb URL, or returns
// the input unchanged if it's already a bare ID. Accepts forms like:
//
//	https://www.imdb.com/title/tt1375666/
//	https://imdb.com/title/tt1375666/?ref_=nv_sr_srsg_0
//	tt1375666
func ExtractIMDbID(input string) (string, error) {
	match := imdbIDPattern.FindString(strings.TrimSpace(input))
	if match == "" {
		return "", fmt.Errorf("could not find a valid IMDb ID (expected format like tt1375666) in: %q", input)
	}
	return match, nil
}

// FindByIMDbID looks up a movie/TV show/episode/person on TMDB using an
// IMDb ID (e.g. "tt1375666"). Use ExtractIMDbID first if you have a full
// IMDb URL rather than a bare ID.
func (c *Client) FindByIMDbID(imdbID string) (*FindResults, error) {
	var out FindResults
	params := map[string]string{"external_source": "imdb_id"}
	path := fmt.Sprintf("/find/%s", url.PathEscape(imdbID))
	if err := c.get(path, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FindByIMDbURL is a convenience wrapper that extracts the IMDb ID from a
// full IMDb URL (or accepts a bare ID) and looks it up on TMDB in one call.
func (c *Client) FindByIMDbURL(imdbURLOrID string) (*FindResults, error) {
	id, err := ExtractIMDbID(imdbURLOrID)
	if err != nil {
		return nil, err
	}
	return c.FindByIMDbID(id)
}

// ---------- Search endpoints ----------

// SearchMovies searches for movies matching the given query string.
func (c *Client) SearchMovies(query string, page int) (*PagedMovies, error) {
	var out PagedMovies
	params := map[string]string{
		"query": query,
		"page":  strconv.Itoa(maxInt(page, 1)),
	}
	if err := c.get("/search/movie", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchTVShows searches for TV shows matching the given query string.
func (c *Client) SearchTVShows(query string, page int) (*PagedTVShows, error) {
	var out PagedTVShows
	params := map[string]string{
		"query": query,
		"page":  strconv.Itoa(maxInt(page, 1)),
	}
	if err := c.get("/search/tv", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchPeople searches for people matching the given query string.
func (c *Client) SearchPeople(query string, page int) (*PagedPeople, error) {
	var out PagedPeople
	params := map[string]string{
		"query": query,
		"page":  strconv.Itoa(maxInt(page, 1)),
	}
	if err := c.get("/search/person", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Detail endpoints ----------

// GetMovie fetches full details for a single movie by TMDB ID.
func (c *Client) GetMovie(movieID int) (*Movie, error) {
	var out Movie
	if err := c.get(fmt.Sprintf("/movie/%d", movieID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMovieCredits fetches full details for a single movie's credits by TMDB ID.
func (c *Client) GetMovieCredits(movieID int) (*Credits, error) {
	var out Credits
	if err := c.get(fmt.Sprintf("/movie/%d/credits", movieID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTVCredits fetches full details for a single movie's credits by TMDB ID.
func (c *Client) GetTVCredits(tvID int) (*Credits, error) {
	var out Credits
	if err := c.get(fmt.Sprintf("/tv/%d/credits", tvID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTVShow fetches full details for a single TV show by TMDB ID.
func (c *Client) GetTVShow(tvID int) (*TVShow, error) {
	var out TVShow
	if err := c.get(fmt.Sprintf("/tv/%d", tvID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPerson fetches full details for a single person by TMDB ID.
func (c *Client) GetPerson(personID int) (*Person, error) {
	var out Person
	if err := c.get(fmt.Sprintf("/person/%d", personID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- List endpoints ----------

// PopularMovies returns the current popular movies list.
func (c *Client) PopularMovies(page int) (*PagedMovies, error) {
	var out PagedMovies
	params := map[string]string{"page": strconv.Itoa(maxInt(page, 1))}
	if err := c.get("/movie/popular", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TopRatedMovies returns the current top-rated movies list.
func (c *Client) TopRatedMovies(page int) (*PagedMovies, error) {
	var out PagedMovies
	params := map[string]string{"page": strconv.Itoa(maxInt(page, 1))}
	if err := c.get("/movie/top_rated", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// NowPlayingMovies returns movies currently playing in theaters.
func (c *Client) NowPlayingMovies(page int) (*PagedMovies, error) {
	var out PagedMovies
	params := map[string]string{"page": strconv.Itoa(maxInt(page, 1))}
	if err := c.get("/movie/now_playing", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpcomingMovies returns movies with upcoming release dates.
func (c *Client) UpcomingMovies(page int) (*PagedMovies, error) {
	var out PagedMovies
	params := map[string]string{"page": strconv.Itoa(maxInt(page, 1))}
	if err := c.get("/movie/upcoming", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TrendingWindow is either "day" or "week".
type TrendingWindow string

const (
	TrendingDay  TrendingWindow = "day"
	TrendingWeek TrendingWindow = "week"
)

// TrendingMovies returns trending movies for the given time window.
func (c *Client) TrendingMovies(window TrendingWindow) (*PagedMovies, error) {
	var out PagedMovies
	if err := c.get(fmt.Sprintf("/trending/movie/%s", window), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Helpers ----------

// PosterURL builds a full image URL for a poster/backdrop/profile path.
// Common sizes for posters: "w92", "w154", "w185", "w342", "w500", "w780", "original".
func PosterURL(path string, size string) string {
	if path == "" {
		return ""
	}
	if size == "" {
		size = "w500"
	}
	return fmt.Sprintf("%s/%s%s", imageBaseURL, size, path)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------- Example usage ----------

//func main() {
//	apiKey, err := getAPIKey()
//	if err != nil {
//		fmt.Println("Error:", err)
//		os.Exit(1)
//	}
//
//	client := NewClient(apiKey)
//
//	// Example: Look up a Movie/TV show from a pasted IMDb link.
//	imdbInput := "https://www.imdb.com/title/tt1375666/" // e.g. Inception
//	result, err := client.FindByIMDbURL(imdbInput)
//	if err != nil {
//		fmt.Println("Error looking up IMDb ID:", err)
//		os.Exit(1)
//	}
//
//	switch {
//	case len(result.MovieResults) > 0:
//		m := result.MovieResults[0]
//		fmt.Println("Found movie:", m.Title)
//		fmt.Println("Release date:", m.ReleaseDate)
//		fmt.Println("Overview:", m.Overview)
//		fmt.Println("Rating:", m.VoteAverage)
//		fmt.Println("Poster:", PosterURL(m.PosterPath, "w500"))
//
//	case len(result.TVResults) > 0:
//		t := result.TVResults[0]
//		fmt.Println("Found TV show:", t.Name)
//		fmt.Println("First air date:", t.FirstAirDate)
//		fmt.Println("Overview:", t.Overview)
//		fmt.Println("Rating:", t.VoteAverage)
//		fmt.Println("Poster:", PosterURL(t.PosterPath, "w500"))
//
//	case len(result.TVEpisodeResults) > 0:
//		e := result.TVEpisodeResults[0]
//		fmt.Printf("Found episode: S%02dE%02d - %s\n", e.SeasonNumber, e.EpisodeNumber, e.Name)
//		fmt.Println("Overview:", e.Overview)
//
//	case len(result.PersonResults) > 0:
//		p := result.PersonResults[0]
//		fmt.Println("Found person:", p.Name)
//
//	default:
//		fmt.Println("No match found on TMDB for that IMDb ID.")
//	}
//}
