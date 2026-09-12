package cmd

import (
	"fmt"
	"strings"

	"movie-tracker/internal/display"
	"movie-tracker/internal/movie"

	"github.com/spf13/cobra"
)

var (
	updateRating    float64
	updateDirectors string
	updateTags      string
	updateTitle     string
)

// updateCmd defines "movie-tracker update <id>". It allows a user to
// update the information present in a movie.
var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update fields on an existing movie",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		id := args[0]

		m, err := store.GetByID(id)
		if err != nil {
			return fmt.Errorf("could not find movie: %w", err)
		}

		// command.Flags().Changed(name) is the key piece here: it
		// tells you whether the user actually TYPED that flag on the
		// command line, as opposed to it just holding its zero/default
		// value. That distinction matters a lot for --updateRating in
		// particular: -1 is BOTH the flag's default value AND a
		// meaningful input ("mark as unwatched"). If we only checked
		// "is updateRating != -1?" we could never tell "user explicitly
		// chose -1" apart from "user didn't pass this flag at all."
		// Checking Changed() instead sidesteps that ambiguity entirely.
		changedAnything := false

		if command.Flags().Changed("updateRating") {
			if err := applyScore(m, updateRating); err != nil {
				return err
			}
			changedAnything = true
		}

		if command.Flags().Changed("updateDirectors") {
			m.Directors = splitAndTrim(updateDirectors, ",")
			changedAnything = true
		}

		if command.Flags().Changed("updateTags") {
			m.Tags = splitAndTrim(updateTags, ",")
			changedAnything = true
		}

		if command.Flags().Changed("updateTitle") {
			m.Title = updateTitle
			changedAnything = true
		}

		if !changedAnything {
			fmt.Println("No fields provided. Use --updateRating, --updateDirectors, --updateTags, or --updateTitle.")
			return nil
		}

		if err := store.Update(m); err != nil {
			return fmt.Errorf("could not save update: %w", err)
		}

		display.PrintSuccess("Movie updated successfully!")
		display.PrintMovieDetail(m)
		return nil
	},
}

// applyScore interprets the --updateRating sentinel values documented
// in the flag's help text: -1 means "mark as unwatched", -2 means
// "watched but unrated", and anything else must be a real 0–10 rating.
// Keeping this logic in its own function (rather than inline in RunE)
// makes it independently readable and testable, and keeps RunE focused
// on "which fields changed" rather than "what does -2 mean".
func applyScore(m *movie.Movie, score float64) error {
	switch {
	case score == -1:
		m.Watched = false
		m.Status = "unwatched"
		m.Rating = -1
	case score == -2:
		m.Watched = true
		m.Status = "unrated"
		m.Rating = -1
	case score < 0 || score > 10:
		return fmt.Errorf("invalid score %.1f: must be 0-10, or -1 (unwatched)/-2 (unrated)", score)
	default:
		m.Watched = true
		m.Status = "watched"
		m.Rating = score
	}
	return nil
}

// splitAndTrim splits a comma-separated string into a clean []string —
// trimming whitespace around each piece and dropping any empty
// entries (so a trailing comma or accidental double-comma doesn't
// leave a stray "" in the Directors slice).
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	updateCmd.Flags().Float64Var(&updateRating, "updateRating", -1, "Give this movie a rating from 0.0 - 10.0. Use score -1 for unwatched and -2 for Unrated")
	updateCmd.Flags().StringVar(&updateDirectors, "updateDirectors", "", "Update the directors of this movie. Connect the directors using a ','. Ex: \"David Leitch, Chad Stahelski\"")
	updateCmd.Flags().StringVar(&updateTags, "updateTags", "", "Update what franchise this movie belongs in.")
	updateCmd.Flags().StringVar(&updateTitle, "updateTitle", "", "Update what the movie title.")

	rootCmd.AddCommand(updateCmd)
}
