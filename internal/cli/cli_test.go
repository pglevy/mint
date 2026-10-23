package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/jsonv"
)

// harness runs the CLI in-process against the fixture token document, with a
// scripted terminal and prompt.
type harness struct {
	t       *testing.T
	stdout  bytes.Buffer
	stderr  bytes.Buffer
	env     Env
	fetches int
	// prompts records each Select call; answers are consumed in order. The
	// answer cancelAnswer simulates the user pressing Escape.
	prompts []promptCall
	answers []string
	netErr  error
}

type promptCall struct {
	message string
	choices []string
}

const cancelAnswer = "\x00cancel"

func newHarness(t *testing.T) *harness {
	t.Helper()
	data, err := os.ReadFile("testdata/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonv.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t}
	h.env = Env{
		Stdout: &h.stdout,
		Stderr: &h.stderr,
		Getenv: func(string) string { return "" },
		Fetch: func() (jsonv.Value, error) {
			h.fetches++
			if h.netErr != nil {
				return nil, h.netErr
			}
			return doc, nil
		},
		Select: func(message string, choices []string) (string, bool, error) {
			h.prompts = append(h.prompts, promptCall{message, choices})
			if len(h.answers) == 0 {
				h.t.Fatalf("unexpected prompt %q", message)
			}
			answer := h.answers[0]
			h.answers = h.answers[1:]
			if answer == cancelAnswer {
				return "", false, nil
			}
			return answer, true, nil
		},
	}
	return h
}

func (h *harness) tty() *harness {
	h.env.StdinIsTTY, h.env.StdoutIsTTY = true, true
	return h
}

func (h *harness) run(args ...string) int {
	h.t.Helper()
	return Run(args, h.env)
}

func (h *harness) mustContain(stream, got string, wants ...string) {
	h.t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			h.t.Errorf("%s does not contain %q:\n%s", stream, want, got)
		}
	}
}

func TestVersion(t *testing.T) {
	h := newHarness(t)
	for _, flag := range []string{"--version", "-v"} {
		h.stdout.Reset()
		if code := h.run(flag); code != 0 || h.stdout.String() != Version+"\n" {
			t.Errorf("%s: code %d, stdout %q", flag, code, h.stdout.String())
		}
	}
	if h.fetches != 0 {
		t.Error("--version should not fetch")
	}
}

func TestHelpDoesNotFetch(t *testing.T) {
	for _, args := range [][]string{{}, {"--help"}, {"list", "--help"}, {"tokens", "-h"}, {"export", "-h", "--json"}} {
		h := newHarness(t)
		if code := h.run(args...); code != 0 {
			t.Errorf("%v: code %d", args, code)
		}
		if h.fetches != 0 || h.stdout.Len() == 0 {
			t.Errorf("%v: fetches=%d stdout=%q", args, h.fetches, h.stdout.String())
		}
	}
}

func TestUnknownSubcommand(t *testing.T) {
	h := newHarness(t)
	if code := h.run("lst"); code != 1 {
		t.Errorf("code = %d", code)
	}
	want := "Error: Unknown subcommand \"lst\".\nValid subcommands: list, tokens, export\n\nDid you mean \"list\"?\n"
	if h.stderr.String() != want {
		t.Errorf("stderr = %q, want %q", h.stderr.String(), want)
	}

	h = newHarness(t)
	h.run("foobar")
	if strings.Contains(h.stderr.String(), "Did you mean") {
		t.Errorf("unexpected suggestion: %q", h.stderr.String())
	}
}

func TestUnknownSubcommandBeatsHelp(t *testing.T) {
	h := newHarness(t)
	if code := h.run("foobar", "--help"); code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
}

func TestInvalidCategoryIsReportedBeforeFetching(t *testing.T) {
	for _, sub := range []string{"list", "tokens", "export"} {
		h := newHarness(t)
		if code := h.run(sub, "--category", "bogus"); code != 1 {
			t.Errorf("%s: code = %d", sub, code)
		}
		h.mustContain(sub+" stderr", h.stderr.String(), `Unknown category "bogus".`, "color, gradient, typography, spacing")
		if h.fetches != 0 {
			t.Errorf("%s fetched before validating", sub)
		}
	}
}

func TestJSONModeAlsoReportsErrorsOnStdout(t *testing.T) {
	h := newHarness(t)
	h.run("--json", "list", "-c", "bogus")
	h.mustContain("stdout", h.stdout.String(), `"error": "Unknown category \"bogus\".\nValid categories: color, gradient, typography, spacing"`)
	h.mustContain("stderr", h.stderr.String(), `Unknown category "bogus".`)
}

func TestQuietDoesNotSuppressErrors(t *testing.T) {
	h := newHarness(t)
	h.run("-q", "list", "-c", "bogus")
	h.mustContain("stderr", h.stderr.String(), `Unknown category "bogus".`)
}

// --- list -------------------------------------------------------------------

func TestListWithoutCategoryNeedsATerminal(t *testing.T) {
	h := newHarness(t)
	if code := h.run("list"); code != 1 {
		t.Errorf("code = %d", code)
	}
	if got := h.stderr.String(); got != "Error: --category is required when stdin is not a TTY\n" {
		t.Errorf("stderr = %q", got)
	}
	if h.stdout.Len() != 0 || h.fetches != 0 {
		t.Errorf("stdout=%q fetches=%d", h.stdout.String(), h.fetches)
	}
}

func TestListMissingCategoryErrorIsSilencedByQuiet(t *testing.T) {
	// The message is written with the quiet-aware stderr writer, as in the
	// original implementation, so --quiet leaves only the exit status.
	h := newHarness(t)
	if code := h.run("list", "--quiet"); code != 1 || h.stderr.Len() != 0 {
		t.Errorf("code=%d stderr=%q", code, h.stderr.String())
	}
}

func TestListJSONWithoutCategoryListsEverything(t *testing.T) {
	h := newHarness(t)
	if code := h.run("list", "--json"); code != 0 {
		t.Fatalf("code = %d", code)
	}
	out := h.stdout.String()
	for _, cat := range []string{`"category": "color"`, `"category": "gradient"`, `"category": "typography"`, `"category": "spacing"`} {
		h.mustContain("stdout", out, cat)
	}
	// Entries have a stable field order.
	h.mustContain("stdout", out, "{\n    \"path\": \"color.red.50\",\n    \"value\": \"#fef2f2\",\n    \"type\": \"color\",\n    \"category\": \"color\"\n  }")
}

func TestListInteractivePromptsForCategory(t *testing.T) {
	h := newHarness(t).tty()
	h.answers = []string{"spacing"}
	if code := h.run("list"); code != 0 {
		t.Fatalf("code = %d, stderr %q", code, h.stderr.String())
	}
	if len(h.prompts) != 1 || h.prompts[0].message != "Select a category:" ||
		strings.Join(h.prompts[0].choices, ",") != "color,gradient,typography,spacing" {
		t.Errorf("prompts = %+v", h.prompts)
	}
	h.mustContain("stdout", h.stdout.String(), "spacing (", "radius.none")
	if strings.Contains(h.stdout.String(), "typography (") {
		t.Error("listed more than the chosen category")
	}
}

func TestListInteractiveJSONDoesNotPrompt(t *testing.T) {
	h := newHarness(t).tty()
	if code := h.run("list", "--json"); code != 0 || len(h.prompts) != 0 {
		t.Errorf("code=%d prompts=%v", code, h.prompts)
	}
}

func TestListInteractiveCancelExitsQuietly(t *testing.T) {
	h := newHarness(t).tty()
	h.answers = []string{cancelAnswer}
	if code := h.run("list"); code != 1 {
		t.Errorf("code = %d", code)
	}
	if h.stdout.Len() != 0 || h.stderr.Len() != 0 || h.fetches != 0 {
		t.Errorf("stdout=%q stderr=%q fetches=%d", h.stdout.String(), h.stderr.String(), h.fetches)
	}
}

func TestListPromptFailureIsReported(t *testing.T) {
	h := newHarness(t).tty()
	h.env.Select = func(string, []string) (string, bool, error) { return "", false, errors.New("not a terminal") }
	if code := h.run("list"); code != 1 {
		t.Errorf("code = %d", code)
	}
	h.mustContain("stderr", h.stderr.String(), "not a terminal", "Run with --debug for more details.")
}

func TestListQuietOmitsCategoryHeader(t *testing.T) {
	h := newHarness(t)
	h.run("list", "-c", "gradient", "-q")
	if strings.Contains(h.stdout.String(), "tokens)") {
		t.Errorf("header printed despite --quiet:\n%s", h.stdout.String())
	}
	h.mustContain("stdout", h.stdout.String(), "header-bg")
}

// --- color ------------------------------------------------------------------

func TestColorFollowsTerminalAndEnvironment(t *testing.T) {
	const cyanHeader = "\x1b[36mcolor ("
	tests := []struct {
		name    string
		tty     bool
		noColor string
		args    []string
		want    bool
	}{
		{"terminal", true, "", nil, true},
		{"piped", false, "", nil, false},
		{"NO_COLOR set", true, "1", nil, false},
		{"NO_COLOR empty is ignored", true, "", nil, true},
		{"--no-color", true, "", []string{"--no-color"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.env.StdoutIsTTY = tt.tty
			h.env.Getenv = func(k string) string {
				if k == "NO_COLOR" {
					return tt.noColor
				}
				return ""
			}
			h.run(append([]string{"list", "-c", "color"}, tt.args...)...)
			if got := strings.Contains(h.stdout.String(), cyanHeader); got != tt.want {
				t.Errorf("colored = %v, want %v\n%q", got, tt.want, h.stdout.String())
			}
		})
	}
}

func TestErrorsAreRedAndSuggestionsDimOnATerminal(t *testing.T) {
	h := newHarness(t).tty()
	h.run("list", "-c", "bogus")
	want := "\x1b[31mUnknown category \"bogus\".\n\x1b[2mValid categories: color, gradient, typography, spacing\x1b[0m\x1b[0m\n"
	if h.stderr.String() != want {
		t.Errorf("stderr = %q\nwant     %q", h.stderr.String(), want)
	}
}

// --- tokens -----------------------------------------------------------------

func TestTokensPromptsForCategoryThenFamily(t *testing.T) {
	h := newHarness(t).tty()
	h.answers = []string{"spacing", "radius"}
	if code := h.run("tokens"); code != 0 {
		t.Fatalf("code = %d, stderr %q", code, h.stderr.String())
	}
	if len(h.prompts) != 2 {
		t.Fatalf("prompts = %+v", h.prompts)
	}
	if got := h.prompts[1]; got.message != `Select a family in "spacing":` || strings.Join(got.choices, ",") != "radius,margin,padding" {
		t.Errorf("family prompt = %+v", got)
	}
	out := h.stdout.String()
	h.mustContain("stdout", out, `"none"`, `"SQUARED"`)
	if strings.Contains(out, `"margin"`) {
		t.Errorf("output includes other families:\n%s", out)
	}
}

func TestTokensWithCategoryStillPromptsForFamilyOnATerminal(t *testing.T) {
	h := newHarness(t).tty()
	h.answers = []string{"red"}
	h.run("tokens", "--category", "color")
	if len(h.prompts) != 1 || h.prompts[0].message != `Select a family in "color":` {
		t.Errorf("prompts = %+v", h.prompts)
	}
}

func TestTokensExplicitFamilySkipsFamilyPrompt(t *testing.T) {
	h := newHarness(t).tty()
	if code := h.run("tokens", "-c", "color", "-f", "black"); code != 0 || len(h.prompts) != 0 {
		t.Errorf("code=%d prompts=%v", code, h.prompts)
	}
	h.mustContain("stdout", h.stdout.String(), `"black": {`, `"#000000"`)
}

func TestTokensFamilyPromptCancel(t *testing.T) {
	h := newHarness(t).tty()
	h.answers = []string{cancelAnswer}
	if code := h.run("tokens", "-c", "color"); code != 1 || h.stdout.Len() != 0 {
		t.Errorf("code=%d stdout=%q", code, h.stdout.String())
	}
}

func TestTokensPipedOutputDumpsEverything(t *testing.T) {
	// stdin is a terminal but stdout is not: no prompting, all categories.
	h := newHarness(t)
	h.env.StdinIsTTY = true
	if code := h.run("tokens"); code != 0 || len(h.prompts) != 0 {
		t.Fatalf("code=%d prompts=%v", code, h.prompts)
	}
	out := h.stdout.String()
	if !strings.HasPrefix(out, "{\n  \"color\": {") || !strings.Contains(out, "\n  \"spacing\": {") {
		t.Errorf("unexpected output:\n%.200s", out)
	}
}

func TestTokensStdoutTerminalWithoutStdinIsAnError(t *testing.T) {
	h := newHarness(t)
	h.env.StdoutIsTTY = true
	if code := h.run("tokens"); code != 1 {
		t.Errorf("code = %d", code)
	}
	h.mustContain("stderr", h.stderr.String(), "--category is required")
}

func TestTokensFamilyWithoutCategoryIsIgnored(t *testing.T) {
	h := newHarness(t)
	h.run("tokens", "--family", "red")
	if !strings.HasPrefix(h.stdout.String(), "{\n  \"color\": {") {
		t.Errorf("expected the full document:\n%.200s", h.stdout.String())
	}
}

func TestTokensEmptyFamilyMeansWholeCategory(t *testing.T) {
	h := newHarness(t)
	h.run("tokens", "-c", "spacing", "--family")
	h.mustContain("stdout", h.stdout.String(), `"radius"`, `"margin"`, `"padding"`)
}

func TestTokensUnknownFamilyIsEmpty(t *testing.T) {
	h := newHarness(t)
	if code := h.run("tokens", "-c", "color", "-f", "nope"); code != 0 || h.stdout.String() != "{}\n" {
		t.Errorf("code=%d stdout=%q", code, h.stdout.String())
	}
}

// --- export -----------------------------------------------------------------

func TestExportToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	h := newHarness(t)
	if code := h.run("export", "-c", "gradient", "-o", path); code != 0 {
		t.Fatalf("code = %d, stderr %q", code, h.stderr.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "{\n  \"gradient\": {") || !strings.HasSuffix(string(data), "}\n") {
		t.Errorf("file contents:\n%s", data)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q", h.stdout.String())
	}
	if got, want := h.stderr.String(), "Wrote tokens to "+path+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestExportToFileQuiet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	h := newHarness(t)
	h.run("export", "--output", path, "-q")
	if h.stderr.Len() != 0 {
		t.Errorf("stderr = %q", h.stderr.String())
	}
	if _, err := os.Stat(path); err != nil {
		t.Error(err)
	}
}

func TestExportOverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	if err := os.WriteFile(path, []byte(strings.Repeat("old", 10000)), 0o644); err != nil {
		t.Fatal(err)
	}
	h := newHarness(t)
	h.run("export", "-c", "gradient", "-o", path)
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "old") {
		t.Error("file was not truncated")
	}
}

func TestExportToUnwritablePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "tokens.json")
	h := newHarness(t)
	if code := h.run("export", "-o", path); code != 1 {
		t.Errorf("code = %d", code)
	}
	h.mustContain("stderr", h.stderr.String(),
		`Cannot write to "`+path+`".`,
		"Check that the directory exists and you have write permissions.")
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q", h.stdout.String())
	}

	h = newHarness(t)
	h.run("export", "-o", path, "--debug")
	h.mustContain("stderr --debug", h.stderr.String(), "FileWriteError: Cannot write to", "caused by:", "no such file or directory")
}

func TestExportEmptyOutputWritesToStdout(t *testing.T) {
	h := newHarness(t)
	h.run("export", "-c", "gradient", "--output")
	h.mustContain("stdout", h.stdout.String(), `"gradient"`)
}

func TestExportPreservesTokenKeyOrder(t *testing.T) {
	h := newHarness(t)
	h.run("export", "-c", "color")
	out := h.stdout.String()
	// Integer-like keys come first, ascending; other keys keep source order.
	i50, i100, i500, i900, iDefault := strings.Index(out, `"50":`), strings.Index(out, `"100":`), strings.Index(out, `"500":`), strings.Index(out, `"900":`), strings.Index(out, `"DEFAULT":`)
	if !(i50 < i100 && i100 < i500 && i500 < i900 && i900 < iDefault) {
		t.Errorf("red scale out of order: 50@%d 100@%d 500@%d 900@%d DEFAULT@%d", i50, i100, i500, i900, iDefault)
	}
}

// --- failures ---------------------------------------------------------------

func TestFetchFailureIsReportedWithItsSuggestion(t *testing.T) {
	h := newHarness(t)
	h.netErr = errs.Fetch("Failed to fetch tokens from X: boom", "Check your network connection and try again.", errors.New("boom"))
	if code := h.run("export"); code != 1 {
		t.Errorf("code = %d", code)
	}
	if got, want := h.stderr.String(), "Failed to fetch tokens from X: boom\nCheck your network connection and try again.\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestUnexpectedErrorsSuggestDebug(t *testing.T) {
	h := newHarness(t)
	h.netErr = errors.New("kaboom")
	h.run("export")
	if got, want := h.stderr.String(), "kaboom\nRun with --debug for more details.\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}

	h = newHarness(t)
	h.netErr = errors.New("kaboom")
	h.run("export", "--debug")
	h.mustContain("stderr", h.stderr.String(), "kaboom\nRun with --debug for more details.\n", "*errors.errorString: kaboom")

	h = newHarness(t)
	h.netErr = errors.New("kaboom")
	h.run("export", "--quiet")
	if got := h.stderr.String(); got != "kaboom\n" {
		t.Errorf("quiet stderr = %q", got)
	}
}

func TestGlobalFlagsAreRemovedFromSubcommandArguments(t *testing.T) {
	// --json is consumed globally, so --category's value is the next word.
	h := newHarness(t)
	if code := h.run("list", "--category", "--json", "gradient"); code != 0 {
		t.Fatalf("code = %d, stderr %q", code, h.stderr.String())
	}
	out := h.stdout.String()
	h.mustContain("stdout", out, `"category": "gradient"`)
	if strings.Contains(out, `"category": "color"`) {
		t.Errorf("category filter was not applied:\n%.300s", out)
	}

	h = newHarness(t)
	if code := h.run("list", "--json", "-q", "--no-color", "--debug", "-c", "gradient"); code != 0 {
		t.Errorf("code = %d, stderr %q", code, h.stderr.String())
	}
}
