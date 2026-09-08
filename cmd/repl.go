// repl.go adds an interactive mode: when the binary is run with no
// subcommand, instead of exiting after printing help, we drop into a
// loop that reads lines from stdin and dispatches them through the
// SAME cobra command tree used for one-shot invocations (e.g.
// `./movie-tracker add "Dune" --year 2021` from a normal shell still
// works exactly as before — this is additive, not a replacement).
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RunInteractive is the REPL loop itself.
func RunInteractive() {
	fmt.Println("movie-tracker interactive mode. Type a command (add, list, remove, search, help) or 'exit' to quit.")

	// bufio.Scanner reads stdin line by line. This is roughly the Go
	// equivalent of wrapping System.in in a BufferedReader and calling
	// readLine() in a loop in Java.
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("movie-tracker> ")

		// scanner.Scan() returns false on EOF (e.g. user pressed
		// Ctrl+D) or a read error, which is our cue to stop looping —
		// similar to readLine() returning null in Java.
		if !scanner.Scan() {
			fmt.Println()
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		args, err := splitArgs(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error parsing input:", err)
			continue
		}
		if len(args) == 0 {
			continue
		}

		// Let unquoted multi-word values work for flags like
		// --updateTitle without requiring the user to wrap them in
		// quotes — see joinMultiWordFlagValues in argparse.go.
		args = joinMultiWordFlagValues(args)

		// Cobra's flag values are "sticky": if you run `add X --year
		// 2020` and then `add Y` (no --year), the old 2020 would
		// otherwise linger on the flag. Since we're reusing the same
		// command tree across many lines instead of a fresh process
		// each time, we reset every flag back to its default before
		// each line is parsed.
		resetFlags(rootCmd)

		// SetArgs tells cobra to parse THIS slice instead of the real
		// os.Args (which is what it reads from in normal, one-shot
		// mode). Execute() then re-runs cobra's normal matching logic
		// against our command tree — same subcommands, same flags —
		// just fed a different source of arguments each time through
		// the loop.
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			// Cobra already prints the error itself by default, so we
			// don't need to print it again here. We just make sure it
			// doesn't kill the loop (unlike Execute() in root.go,
			// which calls os.Exit — that function is only used for
			// the one-shot, non-interactive path).
			continue
		}
	}

	fmt.Println("Goodbye!")
}

// resetFlags walks a command and all its subcommands, resetting every
// flag to its declared default value. pflag.Flag stores the default
// as a string in DefValue; Value.Set(...) re-parses that string back
// into the flag's real value. This is a bit of reflection-flavored
// bookkeeping you wouldn't normally need in a one-shot CLI — it only
// matters because we're reusing the same *cobra.Command instances
// across many "invocations" within one process.
func resetFlags(command *cobra.Command) {
	command.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range command.Commands() {
		resetFlags(sub)
	}
}

// splitArgs tokenizes a line of input the way a shell would for our
// purposes: whitespace-separated, but text inside double quotes is
// kept together as one argument (so titles with spaces, like
// "The Dark Knight", work). This is a simplified version of what a
// real shell does — good enough for our CLI, not meant to handle
// every edge case (nested quotes, escaping, single quotes, etc.).
func splitArgs(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quote in input")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}
