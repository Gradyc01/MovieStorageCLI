package gemini

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PromptMultipleChoice prints a numbered list of options and reads the
// user's pick from stdin, looping until a valid choice is entered.
// Swap this out for a proper TUI (e.g. the same lib used for `list`'s
// arrow-key navigation) if you want it to look nicer than plain stdin.
func PromptMultipleChoice(question string, options []string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(question)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		line = strings.TrimSpace(line)

		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(options) {
			return options[n-1]
		}

		// Also accept the user just typing the option text directly.
		for _, opt := range options {
			if strings.EqualFold(opt, line) {
				return opt
			}
		}

		fmt.Printf("Please enter a number from 1 to %d.\n", len(options))
	}
}
