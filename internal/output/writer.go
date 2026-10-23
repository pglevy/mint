package output

import (
	"fmt"
	"io"

	"github.com/pglevy/mint/internal/jsonv"
)

// Writer centralizes output routing:
//   - primary data goes to stdout
//   - informational messages go to stderr, and --quiet suppresses them
//   - errors always go to stderr, and in --json mode are also written to
//     stdout as {"error": message}
type Writer struct {
	out   io.Writer
	err   io.Writer
	quiet bool
	json  bool
	color Color
}

// NewWriter returns a Writer for the given streams and options.
func NewWriter(out, err io.Writer, quiet, json bool, color Color) *Writer {
	return &Writer{out: out, err: err, quiet: quiet, json: json, color: color}
}

// Stdout writes text and a newline to stdout.
func (w *Writer) Stdout(text string) {
	fmt.Fprintln(w.out, text)
}

// Stderr writes text and a newline to stderr unless --quiet is set.
func (w *Writer) Stderr(text string) {
	if w.quiet {
		return
	}
	fmt.Fprintln(w.err, text)
}

// JSON writes v to stdout as indented JSON.
func (w *Writer) JSON(v jsonv.Value) {
	fmt.Fprintln(w.out, jsonv.Marshal(v, "  "))
}

// Error writes an error message to stderr (even when --quiet) and, in --json
// mode, a structured copy to stdout.
func (w *Writer) Error(message string) {
	fmt.Fprintln(w.err, w.color.Error(message))
	if w.json {
		obj := jsonv.NewObject()
		obj.Set("error", message)
		fmt.Fprintln(w.out, jsonv.Marshal(obj, "  "))
	}
}
