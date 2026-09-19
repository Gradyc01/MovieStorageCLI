package cmd

import (
	"fmt"
	"movie-tracker/internal/actions"
	"movie-tracker/internal/display"

	"github.com/spf13/cobra"
)

var (
	updateRating    float64
	updateDirectors string
	updateTags      string
	updateTitle     string
	updateActors    string
	updateNote      string
)

// updateCmd defines "movie-tracker update <id>". It allows a user to
// update the information present in a movie.
var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update fields on an existing movie",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		id := args[0]
		fields := actions.UpdateFields{}

		if command.Flags().Changed("updateRating") {
			fields.Rating = &updateRating
		}
		if command.Flags().Changed("updateDirectors") {
			d := actions.SplitAndTrim(updateDirectors, ",")
			fields.Directors = &d
		}
		if command.Flags().Changed("updateTags") {
			t := actions.SplitAndTrim(updateTags, ",")
			fields.Tags = &t
		}
		if command.Flags().Changed("updateTitle") {
			fields.Title = &updateTitle
		}
		if command.Flags().Changed("updateActors") {
			a := actions.SplitAndTrim(updateActors, ",")
			fields.Actors = &a
		}
		if command.Flags().Changed("updateNote") {
			fields.Note = &updateNote
		}

		m, err := store.UpdateMovie(id, fields)
		if err != nil {
			if err.Error() == "no fields provided to update" {
				fmt.Println("No fields provided. Use --updateRating, --updateDirectors, --updateNote, --updateTags, --updateActors, or --updateTitle.")
				return nil
			}
			return err
		}

		display.PrintSuccess("Movie updated successfully!")
		display.PrintMovieDetail(m)
		return nil
	},
}

func init() {
	updateCmd.Flags().Float64Var(&updateRating, "updateRating", -1, "Give this movie a rating from 0.0 - 10.0. Use score -1 for unwatched and -2 for Unrated")
	updateCmd.Flags().StringVar(&updateDirectors, "updateDirectors", "", "Update the directors of this movie. Connect the directors using a ','. Ex: \"David Leitch, Chad Stahelski\"")
	updateCmd.Flags().StringVar(&updateTags, "updateTags", "", "Update what franchise this movie belongs in.")
	updateCmd.Flags().StringVar(&updateTitle, "updateTitle", "", "Update what the movie title.")
	updateCmd.Flags().StringVar(&updateActors, "updateActors", "", "Update who the known actors were")
	updateCmd.Flags().StringVar(&updateNote, "updateNote", "", "Update what the note for this movie is")

	rootCmd.AddCommand(updateCmd)
}
