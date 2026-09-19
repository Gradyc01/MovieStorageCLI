// Package display holds formatting/rendering helpers for the CLI —
// kept separate from cmd/ so that "how do I print a list of movies"
// isn't tangled up with "how do I parse command line args". Same
// separation-of-concerns idea as splitting a Java view layer from a
// controller layer.
package display

import (
	"bufio"
	"fmt"
	"movie-tracker/internal/environment"
	"movie-tracker/internal/movie"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// defaultPageSize is used unless MOVIE_TRACKER_PAGE_SIZE is set to a
// valid positive integer. Reading config from an environment variable
// is a common Go pattern — os.Getenv returns "" (the zero value for
// string) if the variable isn't set, which is why we check for that
// explicitly below rather than treating it as an error.
const defaultPageSize = 15
const MovieTrackerPageSizeVar = "MOVIE_TRACKER_PAGE_SIZE"

// pageSize reads MOVIE_TRACKER_PAGE_SIZE from the environment, falling
// back to defaultPageSize if it's unset or not a valid positive number.
func pageSize() int {
	raw, err := environment.GetVariable(MovieTrackerPageSizeVar)
	if err != nil {
		raw = ""
	}
	if raw == "" {
		return defaultPageSize
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultPageSize
	}
	return n
}

// PrintMovieTable drives the interactive, arrow-key-navigable list view. It
// slices `movies` into pages of `pageSize()` entries, redraws the
// current page, then blocks on a single keypress to decide whether to
// move forward, backward, move the selection cursor, open a movie's
// detail view, or exit back to the caller (e.g. the REPL prompt in
// repl.go, or straight back to the shell in one-shot mode).
func PrintMovieTable(movies []*movie.Movie) error {
	size := pageSize()
	totalPages := (len(movies) + size - 1) / size // integer ceiling division

	currentPage := 0
	// selectedIndex is relative to the CURRENT page's slice, not the
	// overall movies slice — it resets to 0 whenever the page changes,
	// same way a cursor jumps back to the top of a freshly-scrolled
	// list in most UIs.
	selectedIndex := 0

	// Decide the input strategy ONCE, up front, rather than re-checking
	// every loop iteration. interactive controls both which reader we
	// use below and what hint text the footer shows.
	interactive := stdinIsTerminal()
	var fallbackReader *bufio.Reader
	if !interactive {
		fallbackReader = bufio.NewReader(os.Stdin)
	}

	for {
		// "\033[H\033[2J" is a raw ANSI escape sequence: move the
		// cursor to the top-left, then clear the screen. This isn't a
		// Go-specific feature — it's a terminal control code that
		// works the same way regardless of language, as long as the
		// terminal supports ANSI escapes (virtually all modern
		// terminals do, including Windows Terminal).
		fmt.Print("\033[H\033[2J")

		start := currentPage * size
		end := start + size
		if end > len(movies) {
			end = len(movies)
		}
		pageMovies := movies[start:end]

		// Clamp selectedIndex defensively in case the last page has
		// fewer entries than the previous one did.
		if selectedIndex >= len(pageMovies) {
			selectedIndex = len(pageMovies) - 1
		}
		if selectedIndex < 0 {
			selectedIndex = 0
		}

		printMovies(pageMovies, selectedIndex)
		PrintPageFooter(currentPage, totalPages, interactive)

		var k key
		var err error
		if interactive {
			k, err = readKey()
		} else {
			k, err = readKeyFallback(fallbackReader)
		}
		if err != nil {
			// If we can't read key input at all (e.g. stdin closed
			// unexpectedly), fail gracefully rather than hanging or
			// looping forever.
			return fmt.Errorf("could not read key input: %w", err)
		}

		switch k {
		case keyRight:
			if currentPage < totalPages-1 {
				currentPage++
				selectedIndex = 0
			}
		case keyLeft:
			if currentPage > 0 {
				currentPage--
				selectedIndex = 0
			}
		case keyUp:
			if selectedIndex > 0 {
				selectedIndex--
			}
		case keyDown:
			if selectedIndex < len(pageMovies)-1 {
				selectedIndex++
			}
		case keyEnter:
			if len(pageMovies) == 0 {
				continue
			}

			action, err := showDetail(pageMovies[selectedIndex], interactive)
			if err != nil {
				return err
			}

			if action == detailQuit {
				return nil
			}
		case keyQuit:
			return nil
			// keyUnknown: just redraw the same page, ignore the keypress.
		}
	}
}

// detailAction reports what happened while the user was inside
// showDetail, so paginate() knows whether to redraw the list (stayed
// inside `list`) or stop entirely (drop back to the REPL prompt or
// shell, wherever `list` was called from).
type detailAction int

const (
	detailBack detailAction = iota
	detailQuit
)

// showDetail clears the screen and shows one movie's full detail view
// (the same view the standalone `get` command produces), then waits
// for a single keypress with two meanings — deliberately the REVERSE
// of what those keys mean in the outer list view:
//
//	q      -> go back to the list (redraw and keep browsing)
//	Enter  -> exit `list` entirely, back to a normal prompt
//
// The idea behind Enter exiting rather than doing something in-place
// is that once you're back at a prompt, the regular commands
// (`remove <id>`, future `update <id>`, etc.) are already available
// and don't need to be reinvented as a separate mini command language
// inside the detail view.
func showDetail(m *movie.Movie, interactive bool) (detailAction, error) {
	fmt.Print("\033[H\033[2J")
	PrintMovieDetail(m)
	fmt.Println()
	fmt.Println("Press 'q' to go back to the list, or Enter to exit to the prompt.")

	for {
		var k key
		var err error
		if interactive {
			k, err = readKey()
		}
		if err != nil {
			return detailBack, fmt.Errorf("could not read key input: %w", err)
		}

		switch k {
		case keyQuit:
			return detailBack, nil
		case keyEnter:
			return detailQuit, nil
			// Any other key: ignore and wait again — no need to redraw,
			// nothing about the screen has changed.
		}
	}
}

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
			color:  func(m *movie.Movie, s string) string { return colorizeTitle(m, s) },
		},
		{
			header: "DIRECTOR/CREATOR",
			get:    func(m *movie.Movie) string { return m.ListDisplay(m.Directors) },
		},
		{
			header: "TAGS",
			get:    func(m *movie.Movie) string { return m.ListDisplay(m.Tags) },
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

func printMovies(movies []*movie.Movie, selectedIndex int) {
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

	printSeparator()

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

	printSeparator()
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

// PrintPageFooter shows "Page X of Y" plus navigation hints below the
// table. Kept separate from printMovies so the pagination loop in
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
