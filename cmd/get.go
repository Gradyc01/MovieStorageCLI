package cmd

import (
	"fmt"

	"movie-tracker/internal/display"

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
		id := args[0]

		m, err := store.GetByID(id)
		if err != nil {
			return fmt.Errorf("could not get movie: %w", err)
		}

		display.PrintMovieDetail(m)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
