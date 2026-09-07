package cmd

import (
	"bufio"
	"fmt"
	"movie-tracker/internal/environment"
	"os"
	"strconv"

	"movie-tracker/internal/display"
	"movie-tracker/internal/movie"

	"github.com/spf13/cobra"
)

// defaultPageSize is used unless MOVIE_TRACKER_PAGE_SIZE is set to a
// valid positive integer. Reading config from an environment variable
// is a common Go pattern — os.Getenv returns "" (the zero value for
// string) if the variable isn't set, which is why we check for that
// explicitly below rather than treating it as an error.
const defaultPageSize = 15
const MovieTrackerPageSizeVar = "MOVIE_TRACKER_PAGE_SIZE"

// listCmd defines "movie-tracker list" — no positional args needed.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked movies",
	Args:  cobra.NoArgs,
	RunE: func(command *cobra.Command, args []string) error {
		movies, err := store.List()
		if err != nil {
			return fmt.Errorf("could not list movies: %w", err)
		}

		if len(movies) == 0 {
			fmt.Println("No movies tracked yet. Add one with `movie-tracker add <title>`.")
			return nil
		}

		return paginate(movies)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

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

// paginate drives the interactive, arrow-key-navigable list view. It
// slices `movies` into pages of `pageSize()` entries, redraws the
// current page, then blocks on a single keypress to decide whether to
// move forward, backward, move the selection cursor, open a movie's
// detail view, or exit back to the caller (e.g. the REPL prompt in
// repl.go, or straight back to the shell in one-shot mode).
func paginate(movies []*movie.Movie) error {
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

		display.PrintMovies(pageMovies, selectedIndex)
		display.PrintPageFooter(currentPage, totalPages, interactive)

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
			if len(pageMovies) > 0 {
				if err := showDetail(pageMovies[selectedIndex], interactive, fallbackReader); err != nil {
					return err
				}
			}
		case keyQuit:
			return nil
			// keyUnknown: just redraw the same page, ignore the keypress.
		}
	}
}

// showDetail clears the screen, shows one movie's full detail view
// (the same view the standalone `get` command produces), then blocks
// until the user presses something to come back — so `list` and `get`
// share one rendering path instead of drifting apart over time.
func showDetail(m *movie.Movie, interactive bool, fallbackReader *bufio.Reader) error {
	fmt.Print("\033[H\033[2J")
	display.PrintMovieDetail(m)
	fmt.Println()

	if interactive {
		fmt.Println("Press any key to return to the list...")
		_, err := readKey()
		return err
	}

	fmt.Print("Press Enter to return to the list: ")
	_, err := fallbackReader.ReadString('\n')
	return err
}
