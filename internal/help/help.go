// Package help renders mint's help text: concise (no arguments), full
// (--help) and per-subcommand. Examples come before flags.
package help

import (
	"fmt"
	"io"
	"strings"
)

type subcommand struct {
	name        string
	description string
	examples    []string
	flags       []string
}

var subcommands = []subcommand{
	{
		name:        "list",
		description: "List design tokens",
		examples:    []string{"mint list --category color"},
		flags:       []string{"--category <color|gradient|typography|spacing>  Filter by token category"},
	},
	{
		name:        "tokens",
		description: "Output raw DTCG token data",
		examples:    []string{"mint tokens --category typography --family heading"},
		flags: []string{
			"--category <color|gradient|typography|spacing>  Filter by token category",
			"--family <name>                        Filter by token family",
		},
	},
	{
		name:        "export",
		description: "Export tokens as DTCG JSON",
		examples:    []string{"mint export --category spacing -o tokens.json"},
		flags: []string{
			"--category <color|gradient|typography|spacing>  Filter by token category",
			"-o, --output <path>                    Write output to a file",
		},
	},
}

var globalFlags = []string{
	"-h, --help       Show help",
	"-v, --version    Show version",
	"    --json       Output as JSON",
	"-q, --quiet      Suppress informational messages",
	"    --no-color   Disable color output",
	"    --debug      Show detailed error information",
}

// indent prefixes every line with two spaces.
func indent(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = "  " + l
	}
	return out
}

func emit(w io.Writer, lines []string) {
	fmt.Fprintln(w, strings.Join(lines, "\n"))
}

// Concise prints the short help shown when mint is run with no arguments.
func Concise(w io.Writer) {
	emit(w, []string{
		"Mint — query and export Sailwind design tokens.",
		"",
		"Examples:",
		"  mint list --category color",
		"  mint export -o tokens.json",
		"",
		"Subcommands: list, tokens, export",
		"",
		"Run mint --help for full usage information.",
	})
}

// Full prints the help shown for a top-level --help.
func Full(w io.Writer) {
	lines := []string{
		"Mint — query and export Sailwind design tokens.",
		"",
		"Examples:",
		"  mint list                          List all tokens",
		"  mint list --category color         List tokens in a category",
		"  mint tokens --category typography  Output raw DTCG data",
		"  mint export -o tokens.json         Export tokens to a file",
		"",
		"Flags:",
	}
	lines = append(lines, indent(globalFlags)...)
	lines = append(lines, "", "Subcommands:")
	for _, cmd := range subcommands {
		lines = append(lines, fmt.Sprintf("  %-12s %s", cmd.name, cmd.description))
	}
	lines = append(lines, "", "Repository: https://github.com/pglevy/mint")
	emit(w, lines)
}

// Subcommand prints the help for one subcommand.
func Subcommand(w io.Writer, name string) {
	var cmd *subcommand
	for i := range subcommands {
		if subcommands[i].name == name {
			cmd = &subcommands[i]
		}
	}
	if cmd == nil {
		fmt.Fprintf(w, "Unknown subcommand: %s\n", name)
		return
	}

	lines := []string{fmt.Sprintf("mint %s — %s", cmd.name, cmd.description), "", "Examples:"}
	lines = append(lines, indent(cmd.examples)...)
	if len(cmd.flags) > 0 {
		lines = append(lines, "", "Flags:")
		lines = append(lines, indent(cmd.flags)...)
	}
	lines = append(lines, "", "Global Flags:")
	lines = append(lines, indent(globalFlags)...)
	emit(w, lines)
}
