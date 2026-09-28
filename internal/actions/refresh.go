package actions

import (
	"fmt"
	"movie-tracker/internal/movie"
	"time"
)

func (store *Store) RefreshMovies(movies []*movie.Movie) error {

	for _, m := range movies {
		if time.Since(m.LastUpdated) >= calculateTimePassedCheck(m) {
			if err := refreshMovie(m); err != nil {
				return err
			}
			if err := store.storage.Update(m); err != nil {
				return err
			}
			fmt.Printf("Synced movie: %s\n", movie.GetValue(m.Title))
			time.Sleep(time.Second)
		}
	}
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

	refreshValue(&m.Title, refreshedMovie.Title)
	refreshValue(&m.Year, refreshedMovie.Year)
	refreshValue(&m.ReleaseDate, refreshedMovie.ReleaseDate)

	refreshListValue(&m.Directors, refreshedMovie.Directors)
	refreshListValue(&m.Genres, refreshedMovie.Genres)
	refreshListValue(&m.Tags, refreshedMovie.Tags)
	refreshListValue(&m.ProductionCountries, refreshedMovie.ProductionCountries)
	refreshListValue(&m.ProductionCompanies, refreshedMovie.ProductionCompanies)
	refreshListValue(&m.SpokenLanguages, refreshedMovie.SpokenLanguages)
	refreshListValue(&m.KnownActors, refreshedMovie.KnownActors)
	return nil
}

func refreshValue[T any](currentValue *movie.Value[T], newValue movie.Value[T]) {
	if currentValue.EntryType == movie.MANUAL {
		return
	}
	currentValue = &newValue
}

func refreshListValue[T any](currentValue *[]movie.Value[T], newValue []movie.Value[T]) {
	var newList []movie.Value[T]

	for _, v := range newValue {
		newList = append(newList, v)
	}

	for _, v := range *currentValue {
		if v.EntryType == movie.MANUAL {
			newList = append(newList, v)
		}
	}
	currentValue = &newList
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
