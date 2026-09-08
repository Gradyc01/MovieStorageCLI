// Package cmd holds every CLI subcommand. In cobra's convention, this
// package name is literally "cmd" — every file here (root.go, add.go,
// remove.go, ...) declares `package cmd` and contributes one command
// to the same tree. Think of it like several small controller classes
// that all get registered with one dispatcher (rootCmd) at startup.
package cmd

import (
	"fmt"
	"os"

	"movie-tracker/internal/storage"

	"github.com/spf13/cobra"
)

// store is a package-level variable holding our one Storage instance,
// shared by every subcommand in this package (add.go, remove.go, etc.
// all reference this same `store`). This is a simple form of dependency
// sharing — in a bigger app you might pass this around explicitly
// instead of using a package-level var, but for a small CLI this is
// idiomatic and keeps things simple.
//
// Notice the declared type is the Storage INTERFACE, not *JSONStorage.
// Every subcommand codes against Storage, so swapping in SQLiteStorage
// later means changing this one line, not any subcommand file.
var store storage.Storage

// rootCmd is the base command — what runs when you type just
// "movie-tracker" with no subcommand. cobra.Command's Use/Short/Long
// fields populate the auto-generated --help text for you, which is
// one of the things you get for free versus hand-rolling with `flag`.
var rootCmd = &cobra.Command{
	Use:   "movie-tracker",
	Short: "A CLI for tracking movies you want to watch or have watched",
	Long: `movie-tracker lets you add, remove, list, and search movies
from the command line, backed by a local JSON file for now.`,
}

// Execute is called from main.go. It's the single entry point that
// kicks off cobra's parsing of os.Args and dispatches to whichever
// subcommand matches. If something goes wrong, cobra returns an error
// rather than panicking or calling os.Exit itself — we handle exiting
// here, similar to how a Java main() might catch an exception at the
// top level and call System.exit(1).
//
// We explicitly pass cobra its arguments via SetArgs (running them
// through joinMultiWordFlagValues first) rather than letting it read
// os.Args on its own — this is what lets flags like --updateTitle
// accept an unquoted multi-word value even when the binary is run
// directly from a real shell, not just from the REPL.
func Execute() {
	rootCmd.SetArgs(joinMultiWordFlagValues(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// init() runs after every package-level var (including rootCmd) is
// fully initialized, so it's safe to reference rootCmd here even
// though RunInteractive (in repl.go) also refers back to rootCmd.
// Doing this same assignment directly inside the var literal above
// would trigger a compile error ("initialization cycle") — Go's
// cycle checker conservatively treats any identifier mentioned inside
// a var's initializer expression as a hard dependency, even when it's
// only used later, inside a closure that hasn't run yet. Assigning
// the field as a statement here, instead of as part of the literal,
// sidesteps that entirely.
func init() {
	store = storage.NewJSONStorage("movies.json")

	// Run only fires when the user invokes the binary with NO
	// subcommand (e.g. just "./movie-tracker"). Cobra dispatches to a
	// specific subcommand's Run/RunE whenever one matches, so this
	// never fires for "./movie-tracker add ...", etc. — one-shot usage
	// from a normal shell keeps working exactly as before.
	rootCmd.Run = func(command *cobra.Command, args []string) {
		RunInteractive()
	}
}
