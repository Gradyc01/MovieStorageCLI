package actions

import (
	"fmt"
	"movie-tracker/internal/api"
	"movie-tracker/internal/movie"
)

func (store *Store) FindMovie(query string, year string, filmType string) ([]*movie.Movie, error) {

	client := api.NewClientUsingEnvVariable()

	var movieList []*movie.Movie
	if filmType == "movie" {
		pagedMovies, err := client.SearchMovies(query, year, 1)
		if err != nil {
			return nil, err
		}

		movieList, err = movie.NewMoviesFromMovieList(pagedMovies)
		if err != nil {
			return nil, err
		}
	} else if filmType == "tv" {
		pagedTVShows, err := client.SearchTVShows(query, year, 1)
		if err != nil {
			return nil, err
		}

		movieList, err = movie.NewMoviesFromTVList(pagedTVShows)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("filmType must be movie or tv")
	}

	return movieList, nil
}
