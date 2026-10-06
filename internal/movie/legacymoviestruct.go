package movie

import "time"

type LegacyMovieVersion1 struct {
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
	Status              string    `json:"status"` //Can be of value watched, unwatched, unrated, shortlist, watching
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

type LegacyMovieVersion2 struct {
	Version             int             `json:"version"`
	ID                  string          `json:"id"`
	Title               Value[string]   `json:"title"`
	Year                Value[int]      `json:"year"`
	ReleaseDate         Value[string]   `json:"release_date"` //Has default value of ""
	Watched             bool            `json:"watched"`      //Has default value of false
	Rating              float64         `json:"rating"`       //Has default value of -1
	Directors           []Value[string] `json:"director"`     //Has default value of ""
	ImdbID              string          `json:"imdb_id"`      //Has default value of ""
	AddedAt             time.Time       `json:"added_at"`
	Status              string          `json:"status"` //Can be of value watched, unwatched, unrated, shortlist, watching
	Genres              []Value[string] `json:"genres"`
	FilmType            string          `json:"film_type"` //Can be of value TV or Movie
	Tags                []Value[string] `json:"tags"`
	ProductionCompanies []Value[string] `json:"production_companies"`
	ProductionCountries []Value[string] `json:"production_countries"`
	SpokenLanguages     []Value[string] `json:"spoken_languages"`
	KnownActors         []Value[string] `json:"known_actors"`
	TmdbID              string          `json:"tmdb_id"`
	Notes               string          `json:"notes"`
	FinishedAt          string          `json:"finished_at"`
	LastUpdated         time.Time       `json:"last_updated"`
	TVSeason            int             `json:"tv_season"` // 0 == No season number (movie), -1 == Legacy TV series (no season tracked)
}
