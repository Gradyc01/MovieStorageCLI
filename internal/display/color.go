// color.go adds ANSI color codes to terminal output. These are raw
// terminal control sequences (the same family as the "\033[H\033[2J"
// clear-screen sequence used in cmd/list.go) — not a Go language
// feature, just bytes that a terminal interprets specially. \033 is
// the ESC character; "\033[31m" means "start rendering red text",
// and "\033[0m" means "reset back to normal."
package display

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// colorsEnabled is decided once, at package load time, by checking
// whether stdout is a real terminal. If output is being redirected —
// piped to a file, `grep`, `less`, etc. — we skip color entirely.
// Otherwise the raw escape codes would show up as garbage characters
// in whatever's consuming the output. This mirrors what tools like
// git and ls do with their "--color=auto" default.
var colorsEnabled = term.IsTerminal(int(os.Stdout.Fd()))

const (
	// Text Formatting
	reset         = "\033[0m"
	bold          = "\033[1m"
	dim           = "\033[2m"
	italic        = "\033[3m"
	underline     = "\033[4m"
	inverse       = "\033[7m"
	hidden        = "\033[8m"
	strikethrough = "\033[9m"

	// Standard Foreground Colors
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	dimYellow = "\033[38;5;136m" // Warm dark yellow / mustard
	darkGold  = "\033[38;5;142m" // Slightly lighter olive-yellow
	darkOlive = "\033[38;5;100m" // Deep dim yellow-green

	// TrueColor RGB (24-bit) - Exact custom control
	rgbDimYellow = "\033[38;2;180;140;0m" // Custom dark gold

	// High Intensity Foreground Colors
	hiBlack   = "\033[90m" // Gray / Dark Gray
	hiRed     = "\033[91m"
	hiGreen   = "\033[92m"
	hiYellow  = "\033[93m"
	hiBlue    = "\033[94m"
	hiMagenta = "\033[95m"
	hiCyan    = "\033[96m"
	hiWhite   = "\033[97m"

	// Standard Background Colors
	bgBlack   = "\033[40m"
	bgRed     = "\033[41m"
	bgGreen   = "\033[42m"
	bgYellow  = "\033[43m"
	bgBlue    = "\033[44m"
	bgMagenta = "\033[45m"
	bgCyan    = "\033[46m"
	bgWhite   = "\033[47m"

	// High Intensity Background Colors
	bgHiBlack   = "\033[100m"
	bgHiRed     = "\033[101m"
	bgHiGreen   = "\033[102m"
	bgHiYellow  = "\033[103m"
	bgHiBlue    = "\033[104m"
	bgHiMagenta = "\033[105m"
	bgHiCyan    = "\033[106m"
	bgHiWhite   = "\033[107m"
)

// colorize wraps s in the given ANSI code and a trailing reset, unless
// colorsEnabled is false, in which case it returns s unchanged.
// Centralizing this check in one function means every call site gets
// the "skip color when not a terminal" behavior for free, rather than
// having to remember to check colorsEnabled everywhere color is used.
func colorize(code, s string) string {
	if !colorsEnabled {
		return s
	}
	return code + s + reset
}

// PrintSuccess prints a short green confirmation message. Kept here
// (rather than letting cmd/ reach into unexported color/style details)
// so all ANSI color logic stays centralized in this package — cmd/
// just asks for the effect it wants ("show a success message"),
// without needing to know it's implemented with escape codes at all.
func PrintSuccess(msg string) {
	fmt.Println(colorize(green, "✔ "+msg))
}
