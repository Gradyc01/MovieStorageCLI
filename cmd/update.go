package cmd

import (
	"fmt"
	"slices"
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
			m.Directors, changedAnything = parseArrayOfChanges(m.Directors, splitAndTrim(updateDirectors, ","))
		}

		if command.Flags().Changed("updateTags") {
			m.Tags, changedAnything = parseArrayOfChanges(m.Tags, splitAndTrim(updateTags, ","))
		}

		if command.Flags().Changed("updateTitle") {
			m.Title = updateTitle
			changedAnything = true
		}

		if command.Flags().Changed("updateActors") {
			m.KnownActors, changedAnything = parseArrayOfChanges(m.KnownActors, splitAndTrim(updateActors, ","))
		}

		if command.Flags().Changed("updateNote") {
			m.Notes = updateNote
			changedAnything = true
		}

		if !changedAnything {
			fmt.Println("No fields provided. Use --updateRating, --updateDirectors, --updateNote, --updateTags, or --updateTitle.")
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
		m.Status = movie.UNWATCHED
		m.Rating = -1
	case score == -2:
		m.Watched = true
		m.Status = movie.UNRATED
		m.Rating = -1
	case score == -3:
		m.Watched = false
		m.Status = movie.SHORTLIST
		m.Rating = -1
	case score == -4:
		m.Watched = false
		m.Status = movie.WATCHING
		m.Rating = -1
	case score < 0 || score > 10:
		return fmt.Errorf("invalid score %.1f: must be 0-10, or -1 (unwatched)/-2 (unrated)/-3 (shortlisted)/-4 (watching)", score)
	default:
		m.Watched = true
		m.Status = movie.WATCHED
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

// parseArrayOfChanges parses through a given arr of strings that begin with either a + or - dictating whether this item
// should be removed or added from the original arr
func parseArrayOfChanges(original []string, changes []string) ([]string, bool) {
	result := original
	for _, change := range changes {
		prefix := change[0]
		str := change[1:]
		if prefix == '+' {
			if !slices.Contains(result, str) {
				result = append(result, str)
			} else {
				fmt.Printf("%s is already in the list and can't be added\n", str)
				return original, false
			}
		} else if prefix == '-' {
			if slices.Contains(result, str) {
				index := slices.Index(result, str)
				result = slices.Delete(result, index, index+1)
			} else {
				fmt.Printf("%s is can not be found in the list and can't be removed\n", str)
				return original, false
			}
		} else {
			return changes, true
		}
	}
	return result, true
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
