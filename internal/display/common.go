package display

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"movie-tracker/internal/movie"
)

// separatorWidth controls how long the decorative "====" lines are.
// A fixed width is simplest; you could later compute this from
// terminal width via term.GetSize if you want it to be responsive.
const separatorWidth = 100

// columnGap is the number of spaces inserted between columns/label
// and value, mirroring the "4, 2" (minwidth, padding) tabwriter setup
// we used to have.
const columnGap = 2

// printSeparator prints one dim "====" divider line. Both the table
// and the detail view bracket their output with this exact line, so
// it lives here instead of being copy-pasted in two files.
func printSeparator() {
	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
}

// padRight right-pads s with spaces up to width, measured in runes
// (not bytes), so it stays correct for non-ASCII director/title names.
func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
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

// orDash returns the string unchanged, or an em-dash placeholder if
// it's empty — keeps empty optional fields from rendering as a blank,
// hard-to-read gap in the table.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// applyRatingColor holds the actual rating->color decision, decoupled
// from whatever text it's given. Both the table (which needs to color
// an already width-padded cell) and the detail view (which needs to
// color the bare, unpadded value) share this single switch so the two
// views can never drift out of sync on what counts as a "good" vs
// "bad" rating.
func applyRatingColor(m *movie.Movie, text string) string {
	switch {
	case m.Status == movie.UNWATCHED:
		return colorize(dimYellow, text)
	case m.Status == movie.SHORTLIST:
		return colorize(bold, colorize(darkGold, text))
	case m.Status == movie.WATCHING:
		return colorize(blue, text)
	case m.Rating < 0:
		return colorize(dim, text)
	case m.Rating >= 10:
		return colorize(limeGreen, colorize(bold, text))
	case m.Rating >= 9:
		return colorize(limeGreen, text)
	case m.Rating >= 7:
		return colorize(green, text)
	case m.Rating >= 5:
		return colorize(yellow, text)
	default:
		return colorize(red, text)
	}
}

// colorizeRating colors an already width-padded RATING cell for the
// list table.
func colorizeRating(m *movie.Movie, padded string) string {
	return applyRatingColor(m, padded)
}

// colorizeTitle colors an alreadyWidth-padded RATING cell
func colorizeTitle(m *movie.Movie, paddedText string) string {
	switch m.Status {
	case movie.UNWATCHED:
		return colorize(bold, colorize(white, paddedText))
	case movie.SHORTLIST:
		return colorize(bold, colorize(magenta, paddedText))
	case movie.WATCHING:
		return colorize(bold, colorize(blue, paddedText))
	default:
		return colorize(white, paddedText)
	}
}

// ratingCell colors the bare (unpadded) rating value. Kept as its own
// function — rather than inlined at call sites — because it reads
// more clearly at the one remaining call site (movieDetailFields'
// Rating column) than an inline closure would.
func ratingCell(m *movie.Movie) string {
	return applyRatingColor(m, m.RatingOrWatched())
}
