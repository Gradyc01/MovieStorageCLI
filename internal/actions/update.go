package actions

import (
	"errors"
	"fmt"
	"movie-tracker/internal/movie"
	"slices"
	"strings"
)

type UpdateFields struct {
	Title     *string
	Directors *[]string
	Tags      *[]string
	Actors    *[]string
	Note      *string
	Rating    *float64
}

// UpdateMovie is what updateCmd's RunE used to do inline.
func (store *Store) UpdateMovie(id string, fields UpdateFields) (*movie.Movie, error) {
	m, err := store.storage.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("could not find movie: %w", err)
	}

	changedAnything := false

	if fields.Rating != nil {
		if err := applyScore(m, *fields.Rating); err != nil {
			return nil, err
		}
		changedAnything = true
	}

	if fields.Directors != nil {
		m.Directors, changedAnything = parseArrayOfChanges(m.Directors, *fields.Directors)
	}

	if fields.Tags != nil {
		m.Tags, changedAnything = parseArrayOfChanges(m.Tags, *fields.Tags)
	}

	if fields.Title != nil {
		m.Title = *fields.Title
		changedAnything = true
	}

	if fields.Actors != nil {
		m.KnownActors, changedAnything = parseArrayOfChanges(m.KnownActors, *fields.Actors)
	}

	if fields.Note != nil {
		m.Notes = *fields.Note
		changedAnything = true
	}

	if !changedAnything {
		return nil, errors.New("no fields provided to update")
	}

	if err := store.storage.Update(m); err != nil {
		return nil, fmt.Errorf("could not save update: %w", err)
	}

	return m, nil
}

// SplitAndTrim splits a comma-separated string into a clean []string —
// trimming whitespace around each piece and dropping any empty
// entries (so a trailing comma or accidental double-comma doesn't
// leave a stray "" in the Directors slice).
func SplitAndTrim(s, sep string) []string {
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
