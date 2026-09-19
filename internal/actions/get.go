package actions

import (
	"fmt"
	"movie-tracker/internal/display"
	"movie-tracker/internal/movie"
)

func (store *Store) GetMovie(id string) (*movie.Movie, error) {
	m, err := store.storage.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("could not get movie: %w", err)
	}

	display.PrintMovieDetail(m)
	return m, nil
}
