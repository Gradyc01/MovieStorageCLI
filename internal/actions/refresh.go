package actions

import (
	"fmt"
	"movie-tracker/internal/movie"
	"time"
)

func (store *Store) RefreshMovies(movies []*movie.Movie, forceRefresh bool) error {
	for _, m := range movies {
		if time.Since(m.LastUpdated) >= calculateTimePassedCheck(m) || forceRefresh {
			if err := store.RefreshMovie(m); err != nil {
				return err
			}
		}
	}
	return nil
}

func (store *Store) RefreshMovie(m *movie.Movie) error {
	if err := refreshMovie(m); err != nil {
		return err
	}
	if err := store.storage.Update(m); err != nil {
		return err
	}
	fmt.Printf("Synced movie: %s\n", movie.GetValue(m.Title))
	time.Sleep(200)
	return nil
}

func refreshMovie(m *movie.Movie) error {
	m.LastUpdated = time.Now()

	var refreshedMovie *movie.Movie
	var err error
	if m.FilmType == "movie" {
		refreshedMovie, err = movie.NewMovieViaImdbLink(m.ImdbID, -1)
	} else if m.FilmType == "tv" && m.TVSeason != -1 {
		refreshedMovie, err = movie.NewShowViaImdbLink(m.ImdbID, m.TVSeason, -1)
	} else {
		fmt.Printf("invalid film type: %s ", m.FilmType)
		return nil
	}

	if err != nil {
		return err
	}

	m.Title = refreshValue(m.Title, refreshedMovie.Title)
	m.Year = refreshValue(m.Year, refreshedMovie.Year)
	m.ReleaseDate = refreshValue(m.ReleaseDate, refreshedMovie.ReleaseDate)

	m.Directors = refreshListValue(m.Directors, refreshedMovie.Directors)
	m.Genres = refreshListValue(m.Genres, refreshedMovie.Genres)
	m.Tags = refreshListValue(m.Tags, refreshedMovie.Tags)
	m.ProductionCountries = refreshListValue(m.ProductionCountries, refreshedMovie.ProductionCountries)
	m.ProductionCompanies = refreshListValue(m.ProductionCompanies, refreshedMovie.ProductionCompanies)
	m.SpokenLanguages = refreshListValue(m.SpokenLanguages, refreshedMovie.SpokenLanguages)
	m.KnownActors = refreshListValue(m.KnownActors, refreshedMovie.KnownActors)

	m.StreamingProviders = refreshListValue(m.StreamingProviders, refreshedMovie.StreamingProviders)
	m.FreeProviders = refreshListValue(m.FreeProviders, refreshedMovie.FreeProviders)
	m.PurchaseProviders = refreshListValue(m.PurchaseProviders, refreshedMovie.PurchaseProviders)
	return nil
}

func refreshValue[T any](currentValue movie.Value[T], newValue movie.Value[T]) movie.Value[T] {
	if currentValue.EntryType == movie.MANUAL {
		return currentValue
	}
	return newValue
}

func refreshListValue[T any](currentValue []movie.Value[T], newValue []movie.Value[T]) []movie.Value[T] {
	var newList []movie.Value[T]

	for _, v := range newValue {
		newList = append(newList, v)
	}

	for _, v := range currentValue {
		if v.EntryType == movie.MANUAL {
			newList = append(newList, v)
		}
	}
	return newList
}

func calculateTimePassedCheck(m *movie.Movie) time.Duration {
	var timePassed = 14 * (24 * time.Hour)

	releaseDate, err := time.Parse(time.DateOnly, movie.GetValue(m.ReleaseDate))
	if err != nil {
		return timePassed
	}
	if releaseDate.Before(time.Now()) {
		timePassed *= 2
	}
	if releaseDate.Year() < time.Now().Year() {
		timePassed *= 2
	}
	if releaseDate.Year()+10 < time.Now().Year() {
		timePassed *= 5
	}
	return timePassed
}
