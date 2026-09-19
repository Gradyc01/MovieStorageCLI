package actions

import (
	"fmt"
	"movie-tracker/internal/display"
	"movie-tracker/internal/movie"
)

func (store *Store) AddMovie(imdbLink string, rating float64, season int) (*movie.Movie, error) {
	var err error
	var m *movie.Movie
	if season != 0 {
		m, err = movie.NewShowViaImdbLink(imdbLink, season, rating)
	} else {
		m, err = movie.NewMovieViaImdbLink(imdbLink, rating)
	}
	if err != nil {
		return nil, err
	}

	if err := store.storage.Add(m); err != nil {
		return nil, fmt.Errorf("could not add movie: %w", err)
	}

	display.PrintSuccess("Movie added successfully!")
	display.PrintMovieDetail(m)
	return m, nil
}
