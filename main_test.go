package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Integration tests: build the real binary and assert on its stdout, stderr
// and exit code. None of these invocations reach the network.

var binary string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "mint-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	binary = filepath.Join(dir, "mint")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

type result struct {
	stdout, stderr string
	code           int
}

func runCLI(t *testing.T, args ...string) result {
	t.Helper()
	cmd := exec.Command(binary, args...) // stdin is the null device
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exitErr.ExitCode()
	}
	return result{stdout.String(), stderr.String(), code}
}

func mustContain(t *testing.T, stream, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("%s does not contain %q:\n%s", stream, want, got)
		}
	}
}

func TestInvalidCategory(t *testing.T) {
	r := runCLI(t, "list", "--category", "bogus")
	if r.code != 1 {
		t.Errorf("exit = %d, want 1", r.code)
	}
	mustContain(t, "stderr", r.stderr, "bogus")
}

func TestHelp(t *testing.T) {
	r := runCLI(t, "--help")
	if r.code != 0 {
		t.Errorf("exit = %d, want 0", r.code)
	}
	mustContain(t, "stdout", r.stdout, "Mint", "Examples")
}

func TestNoArgs(t *testing.T) {
	r := runCLI(t)
	if r.code != 0 {
		t.Errorf("exit = %d, want 0", r.code)
	}
	mustContain(t, "stdout", r.stdout, "mint", "--help")
}

func TestUnknownSubcommand(t *testing.T) {
	r := runCLI(t, "foobar")
	if r.code != 1 {
		t.Errorf("exit = %d, want 1", r.code)
	}
	mustContain(t, "stderr", r.stderr, "Unknown subcommand", "foobar")
	if r.stdout != "" {
		t.Errorf("stdout = %q, want empty", r.stdout)
	}
}

func TestVersion(t *testing.T) {
	r := runCLI(t, "--version")
	if r.code != 0 || r.stdout != "0.2.0\n" || r.stderr != "" {
		t.Errorf("got %+v", r)
	}
}

func TestListNeedsACategoryWhenNotInteractive(t *testing.T) {
	r := runCLI(t, "list")
	if r.code != 1 {
		t.Errorf("exit = %d, want 1", r.code)
	}
	mustContain(t, "stderr", r.stderr, "--category is required")
}
