// Package display holds formatting/rendering helpers for the CLI —
// kept separate from cmd/ so that "how do I print a list of movies"
// isn't tangled up with "how do I parse command line args". Same
// separation-of-concerns idea as splitting a Java view layer from a
// controller layer.
package display

import (
	"bytes"
	"fmt"
	"movie-tracker/internal/movie"
	"strings"
	"text/tabwriter"
)

// separatorWidth controls how long the decorative "====" lines are.
// A fixed width is simplest; you could later compute this from
// terminal width via term.GetSize if you want it to be responsive.
const separatorWidth = 100

func PrintMovies(movies []*movie.Movie, selectedIndex int) {
	if len(movies) == 0 {
		fmt.Println("No movies to display.")
		return
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))

	// IMPORTANT: we write to an in-memory buffer here, NOT directly to
	// os.Stdout. This is the key fix for a subtle bug: tabwriter counts
	// every byte in a cell — including invisible ANSI color codes — when
	// deciding how much padding a cell needs. If we colored the header
	// line before handing it to tabwriter, the color codes would eat
	// into that cell's "padding budget" without contributing any actual
	// visible width, causing tabwriter to under-pad it relative to
	// uncolored data rows. The fix: give tabwriter only PLAIN text,
	// let it finish computing alignment and produce fully-padded
	// output, and only THEN wrap complete, already-aligned lines in
	// color — at that point they're just finished strings, so adding
	// invisible bytes to their very start/end can't disturb anything.
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)

	fmt.Fprintln(w, "IMDB ID\tTITLE\tDIRECTOR/CREATOR\tFRANCHISE\tRELEASE DATE\tRATING")
	fmt.Fprintln(w, "-------\t-----\t----------------\t---------\t------------\t------")

	for _, m := range movies {
		// The RATING/WATCHED column IS colored here, pre-Flush, and
		// that's fine — tabwriter explicitly excludes the LAST cell
		// of each line from alignment (there's nothing after it to
		// align with), so invisible bytes there never affect padding.
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			m.ImdbID,
			m.Title,
			ratingCell(m),
			m.DirectorsDisplay(),
			orDash(m.Franchise),
			orDash(m.ReleaseDate),
		)
	}

	// Flush computes column widths from the plain text above and
	// writes the fully-padded result into buf.
	w.Flush()

	// Now split the finished, correctly-aligned output back into
	// lines so we can color the header/underline, and append a small
	// arrow marker to the selected row. Appending text here — AFTER
	// Flush — is safe regardless of length, since tabwriter is done
	// computing widths; we're just gluing extra characters onto the
	// end of an already-finished line, which can't retroactively
	// change anything about how it was padded.
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	for i, line := range lines {
		switch {
		case i == 0:
			fmt.Println(colorize(bold+cyan, line))
		case i == 1:
			fmt.Println(colorize(dim, line))
		case i-2 == selectedIndex:
			fmt.Println(line + colorize(bold+cyan, "  ←"))
		default:
			fmt.Println(line)
		}
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
}

// ratingCell colors the RATING/WATCHED column based on the movie's
// state: yellow if it hasn't been watched yet, dim if it's been
// watched but never rated, green if it has an actual rating.
func ratingCell(m *movie.Movie) string {
	text := m.RatingOrWatched()
	switch {
	case !m.Watched:
		return colorize(dimYellow, text)
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

// orDash returns the string unchanged, or an em-dash placeholder if
// it's empty — keeps empty optional fields from rendering as a blank,
// hard-to-read gap in the table.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// PrintPageFooter shows "Page X of Y" plus navigation hints below the
// table. Kept separate from PrintMovies so the pagination loop in
// cmd/list.go can redraw just this part if it ever needs to.
// interactive controls which instructions make sense: raw single-key
// arrow presses, or type-a-letter-then-Enter for terminals that can't
// support raw mode.
func PrintPageFooter(currentPage, totalPages int, interactive bool) {
	pageInfo := colorize(bold, fmt.Sprintf("Page %d of %d", currentPage+1, totalPages))
	if interactive {
		hint := colorize(dim, "→/n: next page, ←/p: previous page, ↑: move up on page, ↓ move down on page, q/Enter: quit")
		fmt.Printf("%s  —  %s\n", pageInfo, hint)
		return
	}
	fmt.Println(pageInfo)
}
