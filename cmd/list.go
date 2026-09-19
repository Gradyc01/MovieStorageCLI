package cmd

import (
	"movie-tracker/internal/display"

	"github.com/spf13/cobra"
)

// listCmd defines "movie-tracker list" — no positional args needed.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked movies",
	Args:  cobra.NoArgs,
	RunE: func(command *cobra.Command, args []string) error {
		movies, err := store.ListMovie()
		if err != nil {
			return err
		}

		return display.PrintMovieTable(movies)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
