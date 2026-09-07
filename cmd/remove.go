package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// removeCmd defines "movie-tracker remove <id>".
var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a movie by its ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		id := args[0]

		if err := store.Remove(id); err != nil {
			return fmt.Errorf("could not remove movie: %w", err)
		}

		fmt.Printf("Removed movie with id %q\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
