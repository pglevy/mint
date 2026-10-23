package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGolden replays the invocations in testdata/golden and compares stdout,
// stderr and the exit status. The expected output was captured from the
// original Deno implementation (commit 616f09f) running against
// testdata/tokens.json with stdin and stdout piped, so these files pin the
// Go port to the behaviour it replaced.
//
// Each file reads:
//
//	args: ["list", "-c", "color"]
//	exit: 0
//	--- stdout ---
//	…
//	--- stderr ---
//	…
func TestGolden(t *testing.T) {
	files, err := filepath.Glob("testdata/golden/*.txt")
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files found (err %v)", err)
	}
	for _, file := range files {
		t.Run(strings.TrimSuffix(filepath.Base(file), ".txt"), func(t *testing.T) {
			raw, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			args, wantCode, wantStdout, wantStderr := parseGolden(t, string(raw))

			h := newHarness(t)
			code := h.run(args...)

			if code != wantCode {
				t.Errorf("exit = %d, want %d", code, wantCode)
			}
			if got := h.stdout.String(); got != wantStdout {
				t.Errorf("stdout differs:\n%s", diff(got, wantStdout))
			}
			if got := h.stderr.String(); got != wantStderr {
				t.Errorf("stderr differs:\n%s", diff(got, wantStderr))
			}
		})
	}
}

func parseGolden(t *testing.T, raw string) (args []string, code int, stdout, stderr string) {
	t.Helper()
	header, rest, ok := strings.Cut(raw, "--- stdout ---\n")
	if !ok {
		t.Fatal("missing stdout marker")
	}
	stdout, stderr, ok = strings.Cut(rest, "--- stderr ---\n")
	if !ok {
		t.Fatal("missing stderr marker")
	}
	for _, line := range strings.Split(strings.TrimSpace(header), "\n") {
		switch {
		case strings.HasPrefix(line, "args: "):
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "args: ")), &args); err != nil {
				t.Fatal(err)
			}
		case strings.HasPrefix(line, "exit: "):
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "exit: ")), &code); err != nil {
				t.Fatal(err)
			}
		}
	}
	return args, code, stdout, stderr
}

// diff reports the first differing line of two outputs.
func diff(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			return fmt.Sprintf("line %d:\n  got:  %q\n  want: %q", i+1, gl, wl)
		}
	}
	return "(no line differs)"
}
