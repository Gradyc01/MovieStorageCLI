package cmd

import (
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
		return store.RemoveMovie(args[0])
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
