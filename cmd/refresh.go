package cmd

import (
	"movie-tracker/internal/display"

	"github.com/spf13/cobra"
)

// refreshCmd defines "movie-tracker refresh <id>".
var refreshCmd = &cobra.Command{
	Use:   "refresh <id>",
	Short: "Re-sync's a specific movie using the TMDB database",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		if movies, err := store.ListMovie(); args[0] == "*" && err == nil {
			return store.RefreshMovies(movies, true)
		}
		m, err := store.GetMovie(args[0])
		if err != nil {
			return err
		}
		if err := store.RefreshMovie(m); err != nil {
			return err
		}
		display.PrintMovieDetail(m)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
}
