// Package prompt implements mint's interactive terminal prompt.
package prompt

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI escape sequences for terminal control.
const (
	clearLine  = "\x1b[2K"
	cursorUp   = "\x1b[1A"
	cursorHide = "\x1b[?25l"
	cursorShow = "\x1b[?25h"
	dim        = "\x1b[2m"
	reset      = "\x1b[0m"
	cyan       = "\x1b[36m"
	bold       = "\x1b[1m"
)

// While the terminal is raw, lines end in CRLF because raw mode turns off the
// terminal's newline translation.
const eol = "\r\n"

func render(out io.Writer, message string, choices []string, selected int) {
	fmt.Fprint(out, bold+message+reset+eol)
	for i, choice := range choices {
		if i == selected {
			fmt.Fprint(out, cyan+"  > "+choice+reset+eol)
		} else {
			fmt.Fprint(out, dim+"    "+choice+reset+eol)
		}
	}
}

func clearRendered(out io.Writer, lines int) {
	fmt.Fprint(out, strings.Repeat(cursorUp+clearLine, lines))
}

// Select presents an interactive list on out and reads keystrokes from in,
// which must be a terminal. The arrow keys move the selection and Enter
// confirms it. Escape, Ctrl+C and end of input cancel, reported as ok=false.
func Select(in *os.File, out io.Writer, message string, choices []string) (choice string, ok bool, err error) {
	if len(choices) == 0 {
		return "", false, nil
	}

	fd := int(in.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", false, err
	}

	selected := 0
	totalLines := len(choices) + 1 // +1 for the message line

	fmt.Fprint(out, cursorHide)
	render(out, message, choices, selected)
	defer func() {
		term.Restore(fd, state)
		clearRendered(out, totalLines)
		// Show the selected value as confirmation.
		fmt.Fprint(out, bold+message+reset+" "+cyan+choices[selected]+reset+"\n")
		fmt.Fprint(out, cursorShow)
	}()

	buf := make([]byte, 64)
	for {
		n, err := in.Read(buf)
		if n == 0 && err != nil {
			if err == io.EOF {
				return "", false, nil
			}
			return "", false, err
		}

		switch input := string(buf[:n]); input {
		case "\x03", "\x1b": // Ctrl+C, or Escape on its own (not an arrow sequence)
			return "", false, nil
		case "\r", "\n":
			return choices[selected], true, nil
		case "\x1b[A": // up
			selected = (selected + len(choices) - 1) % len(choices)
			clearRendered(out, totalLines)
			render(out, message, choices, selected)
		case "\x1b[B": // down
			selected = (selected + 1) % len(choices)
			clearRendered(out, totalLines)
			render(out, message, choices, selected)
		}
	}
}
