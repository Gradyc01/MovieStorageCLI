package actions

import (
	"fmt"
	"movie-tracker/internal/movie"
)

// listCmd defines "movie-tracker list" — no positional args needed.

func (store *Store) ListMovie() ([]*movie.Movie, error) {
	movies, err := store.storage.List()
	if err != nil {
		return nil, fmt.Errorf("could not list movies: %w", err)
	}

	if len(movies) == 0 {
		fmt.Println("No movies tracked yet. Add one with `movie-tracker add <title>`.")
		return movies, nil
	}

	return movies, nil
}
