package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// searchCmd defines "movie-tracker search <query>".
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracked movies by title",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		query := args[0]

		results, err := store.Search(query)
		if err != nil {
			return fmt.Errorf("could not search movies: %w", err)
		}

		if len(results) == 0 {
			fmt.Printf("No movies matched %q\n", query)
			return nil
		}

		for _, m := range results {
			fmt.Printf("[%s] %s\n", m.ID, m)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
