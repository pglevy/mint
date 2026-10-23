package errs

import (
	"errors"
	"testing"
)

func TestSuggest(t *testing.T) {
	valid := []string{"list", "tokens", "export"}
	tests := []struct {
		input string
		want  string // "" means no suggestion
	}{
		{"lst", "list"},
		{"lisst", "list"},
		{"tokenz", "tokens"},
		{"exprt", "export"},
		{"List", "list"},
		{"export", "export"},
		{"l", ""}, // distance 3
		{"foobar", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got, ok := Suggest(tt.input, valid)
		if tt.want == "" && ok {
			t.Errorf("Suggest(%q) = %q, want no suggestion", tt.input, got)
		} else if tt.want != "" && (!ok || got != tt.want) {
			t.Errorf("Suggest(%q) = %q, %v; want %q", tt.input, got, ok, tt.want)
		}
	}
}

func TestSuggestPrefersEarlierCandidateOnTie(t *testing.T) {
	if got, _ := Suggest("ab", []string{"ax", "ay"}); got != "ax" {
		t.Errorf("got %q, want ax", got)
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0}, {"abc", "abc", 0}, {"", "abc", 3}, {"abc", "", 3},
		{"kitten", "sitting", 3}, {"list", "lst", 1}, {"é", "e", 1},
	}
	for _, tt := range tests {
		if got := levenshtein(tt.a, tt.b); got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestErrorKindsAndUnwrap(t *testing.T) {
	cause := errors.New("boom")
	tests := []struct {
		err  *Error
		kind string
	}{
		{Fetch("m", "s", cause), "FetchError"},
		{Validation("m", "s"), "ValidationError"},
		{FileWrite("m", "s", cause), "FileWriteError"},
	}
	for _, tt := range tests {
		if tt.err.Kind != tt.kind || tt.err.Error() != "m" || tt.err.Suggestion != "s" {
			t.Errorf("%+v", tt.err)
		}
	}
	if !errors.Is(Fetch("m", "s", cause), cause) {
		t.Error("cause not unwrappable")
	}
}
