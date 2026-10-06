package storage

import "movie-tracker/internal/movie"

type LegacyMovieStorageVersion2 struct {
	Version int                          `json:"version"`
	Movies  []*movie.LegacyMovieVersion2 `json:"movies"`
}
