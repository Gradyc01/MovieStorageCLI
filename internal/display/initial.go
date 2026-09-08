package display

import (
	"fmt"
	"strings"
)

// commandInfo pairs a command name with its one-line description, in
// the exact order they should be listed in the welcome banner.
type commandInfo struct {
	name string
	desc string
}

// PrintInitialWelcomeMessage renders the banner shown when the CLI
// starts with no subcommand, listing every top-level command and what
// it does.
func PrintInitialWelcomeMessage() {
	commands := []commandInfo{
		{"add", "Add a new movie to your tracker"},
		{"get", "Show detailed information about a specific movie"},
		{"remove", "Remove a movie by its ID"},
		{"list", "List all tracked movies"},
		{"search", "Search tracked movies by title"},
		{"update", "Update fields on an existing movie"},
	}

	// Right-pad every command name to the widest one, the same
	// plain-text-width-first approach used in table.go, so the
	// descriptions all start in a clean vertical line regardless of
	// color codes.
	maxName := 0
	for _, c := range commands {
		if n := len(c.name); n > maxName {
			maxName = n
		}
	}

	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
	fmt.Println(colorize(bold+cyan, centerText("MOVIE TRACKER", separatorWidth)))
	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
	fmt.Println()
	fmt.Println(colorize(bold, "Available commands:"))
	fmt.Println()

	for _, c := range commands {
		name := c.name + strings.Repeat(" ", maxName-len(c.name))
		fmt.Printf("  %s  %s\n", colorize(bold+cyan, name), colorize(dim, c.desc))
	}

	fmt.Println()
	fmt.Println(colorize(dim, strings.Repeat("=", separatorWidth)))
}
