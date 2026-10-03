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
	"slices"
	"strconv"
	"strings"
	"time"
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
// detail view, enter sort mode, or exit back to the caller (e.g. the
// REPL prompt in repl.go, or straight back to the shell in one-shot mode).
func PrintMovieTable(movies []*movie.Movie) error {
	size := pageSize()
	totalPages := (len(movies) + size - 1) / size // integer ceiling division

	currentPage := 0
	// selectedIndex is relative to the CURRENT page's slice, not the
	// overall movies slice — it resets to 0 whenever the page changes,
	// same way a cursor jumps back to the top of a freshly-scrolled
	// list in most UIs.
	selectedIndex := 0

	// Sorting state. `movies` is never mutated; `view` is the sorted
	// copy that we actually page over. Sorting doesn't change the
	// number of movies, so totalPages stays valid.
	cols := movieColumns()
	sortable := sortableIndexes(cols)
	st := noSort()
	sortMode := false
	sortCursor := 0 // index into `sortable`, not into cols
	view := movies

	// enterSortMode is shared by the `s` key and by pressing ↑ while
	// the row cursor is already at the top of the page. The sort
	// cursor starts on the currently sorted column, if any.
	enterSortMode := func() {
		if len(sortable) == 0 || len(movies) == 0 {
			return
		}
		sortMode = true
		sortCursor = 0
		if i := slices.Index(sortable, st.col); i >= 0 {
			sortCursor = i
		}
	}

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
		if end > len(view) {
			end = len(view)
		}
		pageMovies := view[start:end]

		if selectedIndex >= len(pageMovies) {
			selectedIndex = len(pageMovies) - 1
		}
		if selectedIndex < 0 {
			selectedIndex = 0
		}

		hv := headerView{sortCol: st.col, desc: st.desc, cursorCol: -1}
		rowCursor := selectedIndex
		if sortMode {
			hv.cursorCol = sortable[sortCursor]
			rowCursor = -1
		}

		printMovies(pageMovies, rowCursor, hv)
		printPageFooter(currentPage, totalPages, interactive, sortMode)

		var k key
		var err error
		if interactive {
			k, err = readKey()
		} else {
			prompt := promptList
			if sortMode {
				prompt = promptSort
			}
			k, err = readKeyFallback(fallbackReader, prompt)
		}
		if err != nil {
			// If we can't read key input at all (e.g. stdin closed
			// unexpectedly), fail gracefully rather than hanging or
			// looping forever.
			return fmt.Errorf("could not read key input: %w", err)
		}

		if sortMode {
			resort := false
			switch k {
			case keyLeft:
				if sortCursor > 0 {
					sortCursor--
				}
			case keyRight:
				if sortCursor < len(sortable)-1 {
					sortCursor++
				}
			case keyUp, keyEnter:
				// Cycle the highlighted column: ascending -> descending -> off.
				st, resort = st.cycle(sortable[sortCursor]), true
			case keyDown, keyQuit, keySort:
				sortMode = false
			}
			if resort {
				view = sortMovies(movies, cols, st)
				currentPage, selectedIndex = 0, 0
			}
			continue
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
			} else {
				enterSortMode()
			}
		case keyDown:
			if selectedIndex < len(pageMovies)-1 {
				selectedIndex++
			}
		case keySort:
			enterSortMode()
		case keyEnter:
			if len(pageMovies) == 0 {
				continue
			}

			action, err := showDetail(pageMovies[selectedIndex], interactive, fallbackReader)
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
//
// fallbackReader is only used (and must be non-nil) when interactive
// is false; previously the non-terminal case never read any input and
// looped forever.
func showDetail(m *movie.Movie, interactive bool, fallbackReader *bufio.Reader) (detailAction, error) {
	fmt.Print("\033[H\033[2J")
	PrintMovieDetail(m)
	fmt.Println()
	if interactive {
		fmt.Println("Press 'q' to go back to the list, or Enter to exit to the prompt.")
	}

	for {
		var k key
		var err error
		if interactive {
			k, err = readKey()
		} else {
			k, err = readKeyFallback(fallbackReader, promptDetail)
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
//
// compare and missing are optional and only matter for sorting:
// a column is sortable if (and only if) compare is non-nil. Rows for
// which missing() is true are always placed last, in both directions.
type column struct {
	header  string
	get     func(m *movie.Movie) string
	color   func(m *movie.Movie, padded string) string
	compare func(a, b *movie.Movie) int // nil = not sortable
	missing func(m *movie.Movie) bool   // optional
}

// headerView tells printMovies how to decorate the header row.
type headerView struct {
	sortCol   int  // column showing the ▲/▼ indicator, or -1
	desc      bool // direction of that sort
	cursorCol int  // column highlighted in sort mode, or -1
}

// validDate reports whether s is a YYYY-MM-DD date. FinishedAt holds
// the literal string "unwatched" for unfinished movies, so a failed
// parse means "no date".
func validDate(s string) bool {
	_, err := time.Parse(time.DateOnly, s)
	return err == nil
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
			get:    func(m *movie.Movie) string { return movie.GetValue(m.Title) },
			color:  func(m *movie.Movie, s string) string { return colorizeTitle(m, s) },
			compare: func(a, b *movie.Movie) int {
				return strings.Compare(
					strings.ToLower(movie.GetValue(a.Title)),
					strings.ToLower(movie.GetValue(b.Title)))
			},
		},
		{
			header: "DIRECTOR/CREATOR",
			get:    func(m *movie.Movie) string { return m.ListDisplay(movie.GetValueList(m.Directors)) },
		},
		{
			header: "TAGS",
			get:    func(m *movie.Movie) string { return m.ListDisplay(movie.GetValueList(m.Tags)) },
			color: func(m *movie.Movie, s string) string {
				if !strings.Contains(s, "—") {
					return colorize(darkOlive, s)
				}
				return colorize(dim, s)
			},
			compare: func(a, b *movie.Movie) int {
				return strings.Compare(
					strings.ToLower(a.ListDisplay(movie.GetValueList(a.Tags))),
					strings.ToLower(b.ListDisplay(movie.GetValueList(b.Tags))))
			},
			missing: func(m *movie.Movie) bool { return len(m.Tags) == 0 },
		},
		{
			header: "RELEASE DATE",
			get:    func(m *movie.Movie) string { return orDash(movie.GetValue(m.ReleaseDate)) },
			// YYYY-MM-DD sorts correctly as plain text.
			compare: func(a, b *movie.Movie) int {
				return strings.Compare(movie.GetValue(a.ReleaseDate), movie.GetValue(b.ReleaseDate))
			},
			missing: func(m *movie.Movie) bool { return movie.GetValue(m.ReleaseDate) == "" },
		},
		{
			// This column shows status, so sorting it orders by status
			// rank first and then by the numeric rating within a status.
			header:  "RATING",
			get:     func(m *movie.Movie) string { return m.RatingOrWatched() },
			color:   func(m *movie.Movie, padded string) string { return colorizeRating(m, padded) },
			compare: compareStatusThenRating,
		},
		{
			header: "FINISHED",
			get:    func(m *movie.Movie) string { return orDash(m.FinishedAt) },
			color:  func(m *movie.Movie, padded string) string { return colorize(dim, padded) },
			compare: func(a, b *movie.Movie) int {
				return strings.Compare(a.FinishedAt, b.FinishedAt)
			},
			missing: func(m *movie.Movie) bool { return !validDate(m.FinishedAt) },
		},
		{
			header: "ADDED",
			get: func(m *movie.Movie) string {
				return m.AddedAt.Format(time.DateOnly)
			},
			color: func(m *movie.Movie, padded string) string { return colorize(dim, padded) },
			compare: func(a, b *movie.Movie) int {
				return a.AddedAt.Compare(b.AddedAt)
			},
		},
	}
}

// printMovies renders one page. selectedIndex < 0 hides the row arrow
// (used while in sort mode). hv controls the header decorations.
func printMovies(movies []*movie.Movie, selectedIndex int, hv headerView) {
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
			row[c] = truncateText(col.get(m), 40)
		}
		plain[i] = row
	}

	labels := make([]string, len(cols))
	for c, col := range cols {
		labels[c] = col.header
		if c == hv.sortCol {
			if hv.desc {
				labels[c] += " ▼"
			} else {
				labels[c] += " ▲"
			}
		}
	}

	// Column width = widest PLAIN cell in that column (header included).
	// Widths never see a color code, so no column's padding can be
	// thrown off by however many columns we decide to color.
	// Sortable columns always reserve 2 extra characters for the arrow
	// so the table doesn't jump around when a sort is applied.
	widths := make([]int, len(cols))
	for c, col := range cols {
		widths[c] = utf8.RuneCountInString(col.header)
		if col.compare != nil {
			widths[c] += 2
		}
	}
	for _, row := range plain {
		for c, cell := range row {
			if w := utf8.RuneCountInString(cell); w > widths[c] {
				widths[c] = w
			}
		}
	}

	printSeparator()

	var hb strings.Builder
	for c := range cols {
		style := bold + cyan
		if c == hv.cursorCol {
			style = bold + inverse
		}
		hb.WriteString(colorize(style, padRight(labels[c], widths[c])))
		if c != len(cols)-1 {
			hb.WriteString(strings.Repeat(" ", columnGap))
		}
	}
	fmt.Println(hb.String())

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
// joins them with columnGap spaces — used for the underline row,
// which gets colored as a whole line rather than per-cell.
func buildPlainRow(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = padRight(cell, widths[i])
	}
	return strings.TrimRight(strings.Join(parts, strings.Repeat(" ", columnGap)), " ")
}

// printPageFooter shows "Page X of Y" plus navigation hints below the
// table. interactive controls which instructions make sense: raw
// single-key arrow presses, or type-a-letter-then-Enter for terminals
// that can't support raw mode (the fallback reader prints its own
// prompt, so no hint is needed there). sortMode swaps in the hints for
// choosing a sort column.
func printPageFooter(currentPage, totalPages int, interactive, sortMode bool) {
	pageInfo := colorize(bold, fmt.Sprintf("Page %d of %d", currentPage+1, totalPages))
	if interactive {
		var text string
		if sortMode {
			text = "SORT: ←/→ choose column, ↑/Enter: cycle ▲ → ▼ → off, ↓/q/s: done"
		} else {
			text = "→/n: next, ←/p: prev, ↑/↓: select, ↑ at top/s: sort, q/Enter: quit"
		}
		fmt.Printf("%s  —  %s\n", pageInfo, colorize(dim, text))
		return
	}
	fmt.Println(pageInfo)
}
