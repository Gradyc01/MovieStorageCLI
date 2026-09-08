package cmd

import (
	"fmt"

	"movie-tracker/internal/display"

	"github.com/spf13/cobra"
)

// removeCmd defines "movie-tracker remove <id>".
var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a movie by its ID",
	Args:  cobra.ExactArgs(1),
	// By default, if RunE returns an error, cobra ALSO prints its own
	// plain "Error: ..." line plus a usage block. Since we're printing
	// our own colored message below, that would show the same failure
	// twice in two different styles. Silencing both here means we
	// still RETURN the error (so the exit code stays non-zero — useful
	// if this command is ever used in a script), we just own how it
	// gets displayed.
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(command *cobra.Command, args []string) error {
		id := args[0]

		// Look the movie up first, purely so the messages below can
		// reference its title, not just its ID — "Removed 'Dune'" reads
		// a lot better than "Removed movie with id dune-2021".
		m, err := store.GetByID(id)
		if err != nil {
			display.PrintError(fmt.Sprintf("No movie found with id %q", id))
			return err
		}

		if err := store.Remove(id); err != nil {
			display.PrintError(fmt.Sprintf("Could not remove %q: %v", m.Title, err))
			return err
		}

		display.PrintSuccess(fmt.Sprintf("Removed %q (%s)", m.Title, m.ID))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
