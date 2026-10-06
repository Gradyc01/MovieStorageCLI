package storage

import (
	"encoding/json"
	"fmt"
	"movie-tracker/internal/movie"
	"time"
)

func TryLoadingVersion1MovieStorage(data []byte) (*MovieStorage, error) {
	var legacyVersion1Movies []*movie.LegacyMovieVersion1

	if err := json.Unmarshal(data, &legacyVersion1Movies); err != nil {
		return nil, fmt.Errorf("parsing storage file: %w", err)
	}

	var currentVersionMovies []*movie.Movie

	for _, m := range legacyVersion1Movies {
		currentVersionMovies = append(currentVersionMovies, ConvertVersion1MovieToVersion2(m))
	}
	return &MovieStorage{
		Version: CurrentStorageVersion,
		Movies:  currentVersionMovies,
	}, nil
}

func ConvertVersion1MovieToVersion2(v1 *movie.LegacyMovieVersion1) *movie.Movie {
	finishedAt := ""

	if v1.Watched {
		finishedAt = v1.AddedAt.Format(time.DateOnly)
	}
	tvSeason := -1
	if v1.FilmType == "movie" {
		tvSeason = 0
	}

	return &movie.Movie{
		Version:             v1.Version,
		ID:                  v1.ID,
		Title:               movie.Value[string]{Value: v1.Title, EntryType: movie.DATABASE},
		Year:                movie.Value[int]{Value: v1.Year, EntryType: movie.DATABASE},
		ReleaseDate:         movie.Value[string]{Value: v1.ReleaseDate, EntryType: movie.DATABASE},
		Watched:             v1.Watched,
		Rating:              v1.Rating,
		Directors:           movie.ConvertArrayToValueList(v1.Directors, movie.DATABASE),
		ImdbID:              v1.ImdbID,
		AddedAt:             v1.AddedAt,
		Status:              v1.Status,
		Genres:              movie.ConvertArrayToValueList(v1.Genres, movie.DATABASE),
		FilmType:            v1.FilmType,
		Tags:                movie.ConvertArrayToValueList(v1.Tags, movie.DATABASE),
		ProductionCompanies: movie.ConvertArrayToValueList(v1.ProductionCompanies, movie.DATABASE),
		ProductionCountries: movie.ConvertArrayToValueList(v1.ProductionCountries, movie.DATABASE),
		SpokenLanguages:     movie.ConvertArrayToValueList(v1.SpokenLanguages, movie.DATABASE),
		KnownActors:         movie.ConvertArrayToValueList(v1.KnownActors, movie.DATABASE),
		TmdbID:              v1.TmdbID,
		Notes:               v1.Notes,
		FinishedAt:          finishedAt,
		LastUpdated:         v1.AddedAt,
		TVSeason:            tvSeason,
	}
}

func FindConvertibleVersion(version int, data []byte) ([]*movie.Movie, error) {
	var list []*movie.Movie
	switch version {
	case 2:
		{
			var out *LegacyMovieStorageVersion2
			if err := json.Unmarshal(data, &out); err != nil {
				return nil, fmt.Errorf("unmarshalling legacy movie version: %w", err)
			}
			for _, m := range out.Movies {
				list = append(list, convertVersion2MovieToLatest(m))
			}
			return list, nil
		}
	}
	return nil, fmt.Errorf("unknown legacy movie version: %d", version)
}

func convertVersion2MovieToLatest(legacyMovie *movie.LegacyMovieVersion2) *movie.Movie {
	return &movie.Movie{
		Version:             legacyMovie.Version,
		ID:                  legacyMovie.ID,
		Title:               legacyMovie.Title,
		Year:                legacyMovie.Year,
		ReleaseDate:         legacyMovie.ReleaseDate,
		Watched:             legacyMovie.Watched,
		Rating:              legacyMovie.Rating,
		Directors:           legacyMovie.Directors,
		ImdbID:              legacyMovie.ImdbID,
		AddedAt:             legacyMovie.AddedAt,
		Status:              legacyMovie.Status,
		Genres:              legacyMovie.Genres,
		FilmType:            legacyMovie.FilmType,
		Tags:                legacyMovie.Tags,
		ProductionCompanies: legacyMovie.ProductionCompanies,
		ProductionCountries: legacyMovie.ProductionCountries,
		SpokenLanguages:     legacyMovie.SpokenLanguages,
		KnownActors:         legacyMovie.KnownActors,
		TmdbID:              legacyMovie.TmdbID,
		Notes:               legacyMovie.Notes,
		FinishedAt:          legacyMovie.FinishedAt,
		LastUpdated:         time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC),
		TVSeason:            legacyMovie.TVSeason,
		StreamingProviders:  make([]movie.Value[string], 0),
		FreeProviders:       make([]movie.Value[string], 0),
		PurchaseProviders:   make([]movie.Value[string], 0),
	}
}
