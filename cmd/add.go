package cmd

import (
	"fmt"

	"movie-tracker/internal/display"
	"movie-tracker/internal/movie"

	"github.com/spf13/cobra"
)

// Flag variables for this subcommand. cobra convention is to declare
// these at package scope (or file scope) and bind them in init() via
// Flags().StringVar/IntVar, then read them inside Run. This is a bit
// like defining @Option fields in a Java CLI framework such as picocli.
var (
	addSeason int
	rating    float64
)

// addCmd defines "movie-tracker add <title> --year 2024".
// Args: cobra.ExactArgs(1) means this command requires exactly one
// positional argument (the title) and will auto-generate an error
// message if the user gives zero or more than one.
var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new movie to your tracker",
	Args:  cobra.ExactArgs(1),
	// RunE (vs Run) lets us return an error instead of handling it
	// ourselves inline. cobra will print it and set a non-zero exit
	// code automatically — similar to letting a checked exception
	// propagate up to a top-level handler in Java.
	RunE: func(command *cobra.Command, args []string) error {
		var err error
		var m *movie.Movie
		if addSeason != 0 {
			m, err = movie.NewShowViaImdbLink(args[0], addSeason, rating)
		} else {
			m, err = movie.NewMovieViaImdbLink(args[0], rating)
		}
		if err != nil {
			return err
		}

		if err := store.Add(m); err != nil {
			return fmt.Errorf("could not add movie: %w", err)
		}

		display.PrintSuccess("Movie added successfully!")
		display.PrintMovieDetail(m)
		return nil
	},
}

// init() registers this subcommand with rootCmd and declares its
// flags. IntVar's signature is (pointer to bind to, flag name, default
// value, help text) — the pointer is why `addYear` above is declared
// at package scope rather than as a local variable.
func init() {
	addCmd.Flags().IntVar(&addSeason, "season", 0, "Season to add to the tracker")
	addCmd.Flags().Float64Var(&rating, "rating", -1, "Rating to add to the tracker")
	rootCmd.AddCommand(addCmd)
}
