// Package output formats and writes mint's terminal output.
package output

import "github.com/pglevy/mint/internal/tokens"

// ANSI escape codes for terminal colors.
const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)

// categoryColors maps token categories to their display colors.
var categoryColors = map[tokens.Category]string{
	tokens.Color:      ansiCyan,
	tokens.Gradient:   ansiGreen,
	tokens.Typography: ansiMagenta,
	tokens.Spacing:    ansiYellow,
}

// Color applies terminal colors when Enabled; otherwise every method returns
// its text unchanged.
type Color struct {
	Enabled bool
}

func (c Color) wrap(code, text string) string {
	if !c.Enabled {
		return text
	}
	return code + text + ansiReset
}

// Category colors text with the category's display color.
func (c Color) Category(text string, cat tokens.Category) string {
	return c.wrap(categoryColors[cat], text)
}

// Error colors text red.
func (c Color) Error(text string) string { return c.wrap(ansiRed, text) }

// Dim dims text.
func (c Color) Dim(text string) string { return c.wrap(ansiDim, text) }

// Bold makes text bold.
func (c Color) Bold(text string) string { return c.wrap(ansiBold, text) }
