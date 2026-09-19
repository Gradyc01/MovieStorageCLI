package actions

import (
	"fmt"
	"movie-tracker/internal/movie"
)

func (store *Store) SearchMovie(query string) ([]*movie.Movie, error) {
	results, err := store.storage.Search(query)
	if err != nil {
		return nil, fmt.Errorf("could not search movies: %w \n Here is the proper syntax for a search "+
			"\n DIRECTOR contains equals "+
			"\n TITLE    contains equals"+
			"\n YEAR     > < =="+
			"\n RELEASE  > <"+
			"\n RATING   > < =="+
			"\n WATCHED  =="+
			"\n TAGS contains equals "+
			"\n GENRE contains equals "+
			"\n FILM_TYPE contains equals "+
			"\n PROD_COMPANY contains equals "+
			"\n PROD_COUNTRY contains equals "+
			"\n LANGUAGE contains equals "+
			"\n STATUS contains equals "+
			"\n ACTOR contains equals "+
			"\n Example Query: TITLE contains Guardians of, RATING < 9 ", err)
	}

	if len(results) == 0 {
		fmt.Printf("No movies matched %q\n", query)
		return results, nil
	}

	return results, nil
}
