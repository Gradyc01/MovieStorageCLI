package cmd

import (
	"movie-tracker/internal/display"
	"strings"

	"github.com/spf13/cobra"
)

// searchCmd defines "movie-tracker search <query>".
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracked movies by title",
	RunE: func(command *cobra.Command, args []string) error {
		results, err := store.SearchMovie(strings.Join(args, " "))

		if err != nil {
			return err
		}

		return display.PrintMovieTable(results)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
