package display

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"movie-tracker/internal/movie"
)

// PrintMovieDetail renders a single movie as a vertical, labeled
// block — decorative separator lines, a centered title, then one
// "Label:  Value" line per field, with colons aligned. This is
// intentionally a general-purpose renderer: it's used right now for
// the confirmation after `add`, but nothing about it assumes that —
// it would work identically as the output of a future
// `movie-tracker show <id>` command.
func PrintMovieDetail(m *movie.Movie) {
	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
	fmt.Println(colorize(bold+cyan, centerText("MOVIE DETAILS", separatorWidth)))
	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))

	// Same technique as the list table: build the label/value grid as
	// PLAIN text through tabwriter first (so colon alignment is based
	// purely on real, visible label lengths), then apply any color to
	// the VALUE half of each line afterward, once tabwriter is done
	// and the padding is already finalized.
	type field struct {
		label string
		value string
	}
	fields := []field{
		{"ID", m.ID},
		{"Title", m.Title},
		{"Release Date", m.ReleaseDate},
		{"Rating", m.RatingOrWatched()},
		{"Franchise", orDash(m.Franchise)},
		{"Directors", m.DirectorsDisplay()},
		{"IMDB-ID", m.ImdbID},
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)
	for _, f := range fields {
		// Value is the LAST (and only other) cell on its line, so —
		// same rule as before — tabwriter never pads after it. That
		// means we could safely color it even before Flush, but we
		// keep the "color after Flush" habit here anyway for
		// consistency and because Status gets colored conditionally
		// based on the movie's state, which is easier to reason about
		// as a distinct post-processing step.
		fmt.Fprintf(w, "%s:\t%s\n", f.label, f.value)
	}
	w.Flush()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	for i, line := range lines {
		if fields[i].label == "Rating" {
			line = recolorStatusLine(line, m)
		}
		fmt.Println(line)
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
}

// recolorStatusLine finds the already-padded "Status:   <value>" line
// tabwriter produced and wraps just the value portion in color. We
// split on the FIRST tab-turned-spaces boundary isn't directly
// available post-Flush (tabs became spaces), so instead we simply
// re-derive the value text from the movie and re-append it colored,
// reusing the same label prefix tabwriter already aligned for us.
func recolorStatusLine(line string, m *movie.Movie) string {
	plainValue := m.RatingOrWatched()
	idx := strings.LastIndex(line, plainValue)
	if idx == -1 {
		// Fallback: shouldn't happen, but never crash display code
		// over a cosmetic recoloring step.
		return line
	}
	prefix := line[:idx]
	return prefix + ratingCell(m)
}

// centerText pads s with spaces on both sides so it appears centered
// within the given width. If s is already as wide as (or wider than)
// width, it's returned unchanged rather than truncated — decorative
// text overflowing slightly is harmless; cutting off letters isn't.
func centerText(s string, width int) string {
	if len(s) >= width {
		return s
	}
	totalPadding := width - len(s)
	left := totalPadding / 2
	right := totalPadding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
