package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// searchCmd defines "movie-tracker search <query>".
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracked movies by title",
	RunE: func(command *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		results, err := store.Search(query)
		if err != nil {
			return fmt.Errorf("could not search movies: %w \n Here is the proper syntax for a search "+
				"\n DIRECTOR contains equals "+
				"\n TITLE    contains equals"+
				"\n YEAR     > < =="+
				"\n RELEASE  > <"+
				"\n RATING   > < =="+
				"\n WATCHED  =="+
				"\n TAGS contains equals "+
				"\n Example Query: TITLE contains Guardians of, RATING < 9 ", err)
		}

		if len(results) == 0 {
			fmt.Printf("No movies matched %q\n", query)
			return nil
		}

		//display.PrintMovies(results, 0)
		err = paginate(results)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
