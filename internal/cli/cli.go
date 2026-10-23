// Package cli implements the mint command line: global flag handling,
// subcommand routing and top-level error reporting.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/pglevy/mint/internal/args"
	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/help"
	"github.com/pglevy/mint/internal/jsonv"
	"github.com/pglevy/mint/internal/output"
	"github.com/pglevy/mint/internal/prompt"
	"github.com/pglevy/mint/internal/tokens"
	"golang.org/x/term"
)

// Version is the mint version. Release builds can override it with
// -ldflags "-X github.com/pglevy/mint/internal/cli.Version=…".
var Version = "0.2.0"

// Env is everything the commands need from the outside world, so tests can
// substitute it.
type Env struct {
	Stdout      io.Writer
	Stderr      io.Writer
	StdinIsTTY  bool
	StdoutIsTTY bool
	Getenv      func(string) string

	// Fetch returns the parsed token document.
	Fetch func() (jsonv.Value, error)
	// Select asks the user to pick one of choices; ok is false if they cancel.
	Select func(message string, choices []string) (choice string, ok bool, err error)
}

// OSEnv returns the Env for the real process.
func OSEnv() Env {
	return Env{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		StdinIsTTY:  term.IsTerminal(int(os.Stdin.Fd())),
		StdoutIsTTY: term.IsTerminal(int(os.Stdout.Fd())),
		Getenv:      os.Getenv,
		Fetch: func() (jsonv.Value, error) {
			return tokens.Fetch(context.Background(), tokens.URL)
		},
		Select: func(message string, choices []string) (string, bool, error) {
			return prompt.Select(os.Stdin, os.Stderr, message, choices)
		},
	}
}

// interactive reports whether both stdin and stdout are terminals, so
// prompting is possible.
func (e Env) interactive() bool { return e.StdinIsTTY && e.StdoutIsTTY }

// globals are the options accepted by every subcommand.
type globals struct {
	json    bool
	quiet   bool
	noColor bool
	debug   bool
	help    bool
	version bool
}

// session is what a subcommand handler runs with.
type session struct {
	env     Env
	globals globals
	color   output.Color
	out     *output.Writer
}

// exitError asks Run to exit with a status, without printing anything more.
type exitError int

func (e exitError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

type handler func(s *session, argv []string) error

var handlers = map[string]handler{
	"list":   listCommand,
	"tokens": tokensCommand,
	"export": exportCommand,
}

var validSubcommands = []string{"list", "tokens", "export"}

var globalFlagArgs = []string{
	"--json", "--quiet", "-q", "--no-color", "--debug", "--help", "-h", "--version", "-v",
}

// Run executes mint with the given arguments (excluding the program name) and
// returns the process exit status.
func Run(argv []string, env Env) int {
	err := run(argv, env)
	if err == nil {
		return 0
	}
	var exit exitError
	if errors.As(err, &exit) {
		return int(exit)
	}
	reportError(argv, env, err)
	return 1
}

func newContext(env Env, g globals) *session {
	color := output.Color{
		Enabled: env.StdoutIsTTY && env.Getenv("NO_COLOR") == "" && !g.noColor,
	}
	return &session{
		env:     env,
		globals: g,
		color:   color,
		out:     output.NewWriter(env.Stdout, env.Stderr, g.quiet, g.json, color),
	}
}

func run(argv []string, env Env) error {
	parsed := args.Parse(argv, args.Options{
		Bool:  []string{"help", "version", "json", "quiet", "no-color", "debug"},
		Alias: map[string]string{"h": "help", "v": "version", "q": "quiet"},
	})
	g := globals{
		help:    parsed.Truthy("help"),
		version: parsed.Truthy("version"),
		json:    parsed.Truthy("json"),
		quiet:   parsed.Truthy("quiet"),
		noColor: parsed.Truthy("no-color"),
		debug:   parsed.Truthy("debug"),
	}
	c := newContext(env, g)

	if g.version {
		c.out.Stdout(Version)
		return nil
	}

	positionals := parsed.Positionals
	if len(positionals) == 0 {
		if g.help {
			help.Full(env.Stdout)
		} else {
			help.Concise(env.Stdout)
		}
		return nil
	}

	subcommand := positionals[0]

	// The handler gets the raw arguments after the subcommand name, minus the
	// global flags that were already consumed.
	var subArgs []string
	if idx := slices.Index(argv, subcommand); idx >= 0 {
		for _, a := range argv[idx+1:] {
			if !slices.Contains(globalFlagArgs, a) {
				subArgs = append(subArgs, a)
			}
		}
	}

	if !slices.Contains(validSubcommands, subcommand) {
		message := fmt.Sprintf(`Error: Unknown subcommand "%s".`, subcommand)
		message += "\nValid subcommands: " + strings.Join(validSubcommands, ", ")
		if suggestion, ok := errs.Suggest(subcommand, validSubcommands); ok {
			message += fmt.Sprintf("\n\nDid you mean \"%s\"?", suggestion)
		}
		c.out.Error(message)
		return exitError(1)
	}

	if g.help {
		help.Subcommand(env.Stdout, subcommand)
		return nil
	}

	return handlers[subcommand](c, subArgs)
}

// reportError prints a top-level failure. It re-parses the global flags so
// that errors are formatted correctly even if the failure happened before the
// options were known.
func reportError(argv []string, env Env, err error) {
	parsed := args.Parse(argv, args.Options{
		Bool:  []string{"json", "no-color", "debug", "quiet"},
		Alias: map[string]string{"q": "quiet"},
	})
	g := globals{
		json:    parsed.Truthy("json"),
		noColor: parsed.Truthy("no-color"),
		debug:   parsed.Truthy("debug"),
		quiet:   parsed.Truthy("quiet"),
	}
	c := newContext(env, g)

	var mintErr *errs.Error
	if errors.As(err, &mintErr) {
		message := mintErr.Message
		if mintErr.Suggestion != "" {
			message += "\n" + c.color.Dim(mintErr.Suggestion)
		}
		c.out.Error(message)
		if g.debug {
			c.out.Stderr(debugDetail(mintErr.Kind, mintErr))
		}
		return
	}

	c.out.Error(err.Error())
	c.out.Stderr("Run with --debug for more details.")
	if g.debug {
		c.out.Stderr(debugDetail(fmt.Sprintf("%T", err), err))
	}
}

// debugDetail renders the failure class, message and cause chain for --debug.
func debugDetail(kind string, err error) string {
	detail := kind + ": " + err.Error()
	for cause := errors.Unwrap(err); cause != nil; cause = errors.Unwrap(cause) {
		detail += "\n  caused by: " + cause.Error()
	}
	return detail
}
