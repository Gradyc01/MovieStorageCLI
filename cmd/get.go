package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd defines "movie-tracker get <id>". It's the standalone,
// one-shot version of what pressing Enter on a row inside `list` also
// triggers — both paths end up calling the same display.PrintMovieDetail,
// so the two features stay visually consistent for free.
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show detailed information about a specific movie",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		_, err := store.GetMovie(args[0])
		return err
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
