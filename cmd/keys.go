// keys.go reads single keypresses (including arrow keys) from the
// terminal, instead of waiting for the user to press Enter.
//
// Why this needs a library at all: a terminal normally runs in
// "cooked" (a.k.a. canonical) mode, where the OS buffers your typing
// line-by-line and only hands it to the program once you press Enter
// — that's what bufio.Scanner relies on in repl.go. To catch a single
// arrow-key press immediately, the terminal has to be switched into
// "raw" mode, which is OS-level terminal configuration that Go's
// standard library doesn't expose. golang.org/x/term (an official
// Go team package, just not bundled into the stdlib) wraps the
// platform-specific syscalls for both Unix and Windows.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// key is a small enum of the keypresses we care about. Go doesn't
// have a dedicated `enum` keyword like Java — the idiomatic pattern
// is a named int type plus `iota`, which auto-increments starting
// at 0 for each constant in the block.
type key int

const (
	keyUnknown key = iota
	keyLeft
	keyRight
	keyUp
	keyDown
	keyEnter
	keyQuit
)

// stdinIsTerminal reports whether os.Stdin is a real, raw-mode-capable
// console/terminal. Some environments — Git Bash / MSYS2 on Windows,
// certain IDE-integrated terminals, piped input, CI runners — present
// stdin as something else (often a pipe), and asking those to switch
// into raw mode fails outright rather than degrading gracefully. We
// check this once up front so we can pick the right input strategy
// instead of assuming raw mode always works.
func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// readKey puts stdin into raw mode just long enough to read one
// keypress, then restores the terminal to its normal (cooked) mode
// before returning — via `defer term.Restore(...)`, which guarantees
// that cleanup runs even if we return early or something panics.
// Leaving the terminal stuck in raw mode would make the user's shell
// behave strangely after the program exits, so this restore step is
// not optional.
//
// Only call this after confirming stdinIsTerminal() — otherwise
// term.MakeRaw fails (e.g. Windows' "The handle is invalid" when
// stdin isn't a real console).
func readKey() (key, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return keyUnknown, err
	}
	defer term.Restore(fd, oldState)

	// Arrow keys are sent as a 3-byte escape sequence: ESC, '[', then
	// a letter (A=up, B=down, C=right, D=left). A plain letter key is
	// just 1 byte. We read up to 3 bytes and figure out which case we
	// got based on how many bytes actually came back.
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return keyUnknown, err
	}

	if n == 1 {
		switch buf[0] {
		case 'q', 'Q', 27: // 27 = Esc pressed alone
			return keyQuit, nil
		case 13, 10: // 13 = CR (Enter on most terminals), 10 = LF
			return keyEnter, nil
		case 'n', 'N':
			return keyRight, nil
		case 'p', 'P':
			return keyLeft, nil
		}
		return keyUnknown, nil
	}

	if n == 3 && buf[0] == 27 && buf[1] == '[' {
		switch buf[2] {
		case 'C':
			return keyRight, nil
		case 'D':
			return keyLeft, nil
		case 'A':
			return keyUp, nil
		case 'B':
			return keyDown, nil
		}
	}

	return keyUnknown, nil
}

// readKeyFallback is used when stdinIsTerminal() is false. It can't
// react to a single keypress instantly — it has to wait for a full
// line — so it prompts the user to type a letter and press Enter
// instead. Functionally equivalent to readKey() from the caller's
// perspective (same key type returned), just a different input UX
// for environments that can't support raw mode. Since there's no way
// to "hold the arrow key down" without raw mode, up/down selection is
// mapped to letters (u/d) instead, and pressing Enter with nothing
// typed selects whatever row is currently highlighted.
func readKeyFallback(reader *bufio.Reader) (key, error) {
	fmt.Print("Enter n/p (page), u/d (select), Enter (view), or q (quit): ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return keyUnknown, err
	}
	line = strings.TrimSpace(strings.ToLower(line))

	switch line {
	case "n":
		return keyRight, nil
	case "p":
		return keyLeft, nil
	case "u":
		return keyUp, nil
	case "d":
		return keyDown, nil
	case "":
		return keyEnter, nil
	case "q":
		return keyQuit, nil
	}
	return keyUnknown, nil
}
