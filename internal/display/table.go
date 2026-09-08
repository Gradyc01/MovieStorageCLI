// Package display holds formatting/rendering helpers for the CLI —
// kept separate from cmd/ so that "how do I print a list of movies"
// isn't tangled up with "how do I parse command line args". Same
// separation-of-concerns idea as splitting a Java view layer from a
// controller layer.
package display

import (
	"fmt"
	"movie-tracker/internal/movie"
	"strings"
	"unicode/utf8"
)

// separatorWidth controls how long the decorative "====" lines are.
// A fixed width is simplest; you could later compute this from
// terminal width via term.GetSize if you want it to be responsive.
const separatorWidth = 100

// columnGap is the number of spaces inserted between columns, mirroring
// the "4, 2" (minwidth, padding) tabwriter setup we used to have.
const columnGap = 2

// column describes one field of the table. get() must return PLAIN
// text (no ANSI codes) — that's what column widths are computed
// from. color(), if non-nil, is applied AFTER the plain text has
// already been padded to the column's width, so it's free to color
// any column, not just the last one.
type column struct {
	header string
	get    func(m *movie.Movie) string
	color  func(m *movie.Movie, padded string) string
}

func movieColumns() []column {
	return []column{
		{
			header: "IMDB ID",
			get:    func(m *movie.Movie) string { return m.ImdbID },
			color:  func(m *movie.Movie, s string) string { return colorize(dim, s) },
		},
		{
			header: "TITLE",
			get:    func(m *movie.Movie) string { return m.Title },
			color:  func(m *movie.Movie, s string) string { return colorize(white, s) },
		},
		{
			header: "DIRECTOR/CREATOR",
			get:    func(m *movie.Movie) string { return m.DirectorsDisplay() },
		},
		{
			header: "FRANCHISE",
			get:    func(m *movie.Movie) string { return orDash(m.Franchise) },
			color: func(m *movie.Movie, s string) string {
				if !strings.Contains(s, "—") {
					return colorize(darkOlive, s)
				}
				return colorize(dim, s)
			},
		},
		{header: "RELEASE DATE", get: func(m *movie.Movie) string { return orDash(m.ReleaseDate) }},
		{
			header: "RATING",
			get:    func(m *movie.Movie) string { return m.RatingOrWatched() },
			color:  func(m *movie.Movie, padded string) string { return colorizeRating(m, padded) },
		},
	}
}

func PrintMovies(movies []*movie.Movie, selectedIndex int) {
	if len(movies) == 0 {
		fmt.Println("No movies to display.")
		return
	}

	cols := movieColumns()

	// Pre-compute every cell's PLAIN text once. We need it twice (once
	// to measure widths, once to render), and get() may not be cheap
	// (e.g. DirectorsDisplay joins a slice).
	plain := make([][]string, len(movies))
	for i, m := range movies {
		row := make([]string, len(cols))
		for c, col := range cols {
			row[c] = col.get(m)
		}
		plain[i] = row
	}

	// Column width = widest PLAIN cell in that column (header included).
	// This is the whole fix: widths never see a color code, so no
	// column's padding can be thrown off by however many columns we
	// decide to color.
	widths := make([]int, len(cols))
	for c, col := range cols {
		widths[c] = utf8.RuneCountInString(col.header)
	}
	for _, row := range plain {
		for c, cell := range row {
			if w := utf8.RuneCountInString(cell); w > widths[c] {
				widths[c] = w
			}
		}
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))

	headerLine := buildPlainRow(headersOf(cols), widths)
	fmt.Println(colorize(bold+cyan, headerLine))

	underlineCells := make([]string, len(cols))
	for c, col := range cols {
		underlineCells[c] = strings.Repeat("-", utf8.RuneCountInString(col.header))
	}
	fmt.Println(colorize(dim, buildPlainRow(underlineCells, widths)))

	for i, m := range movies {
		var b strings.Builder
		for c, col := range cols {
			padded := padRight(plain[i][c], widths[c])
			cell := padded
			if col.color != nil {
				// Coloring happens AFTER padding, so the padding
				// spaces ride along inside the color codes. That's
				// harmless — a space has no visible foreground — and
				// it means color never touches the width math above.
				cell = col.color(m, padded)
			}
			b.WriteString(cell)
			if c != len(cols)-1 {
				b.WriteString(strings.Repeat(" ", columnGap))
			}
		}
		line := strings.TrimRight(b.String(), " ")
		if i == selectedIndex {
			line += colorize(bold+cyan, "  ←")
		}
		fmt.Println(line)
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
}

// buildPlainRow pads a row of plain strings to the given widths and
// joins them with columnGap spaces — used for the header and
// underline rows, which get colored as a whole line rather than
// per-cell.
func buildPlainRow(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = padRight(cell, widths[i])
	}
	return strings.TrimRight(strings.Join(parts, strings.Repeat(" ", columnGap)), " ")
}

func headersOf(cols []column) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.header
	}
	return out
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

// applyRatingColor holds the actual rating->color decision, decoupled
// from whatever text it's given. Both the table (which needs to color
// an already width-padded cell) and detail.go (which needs to color
// the bare, unpadded value via ratingCell) share this single switch
// so the two views can never drift out of sync on what counts as a
// "good" vs "bad" rating.
func applyRatingColor(m *movie.Movie, text string) string {
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

// colorizeRating colors an already width-padded RATING cell for the
// list table.
func colorizeRating(m *movie.Movie, padded string) string {
	return applyRatingColor(m, padded)
}

// ratingCell colors the bare (unpadded) rating value. Kept as its own
// function — rather than inlined at call sites — because detail.go's
// PrintMovieDetail calls it directly to recolor its "Rating:" line.
func ratingCell(m *movie.Movie) string {
	return applyRatingColor(m, m.RatingOrWatched())
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
