package args

import (
	"slices"
	"testing"
)

// The expectations mirror the behaviour of @std/cli parseArgs 1.0.29, which
// the original Deno implementation used. They were cross-checked against it
// (including a 40,000-vector randomized comparison when this package was
// written).

var mainOpts = Options{
	Bool:  []string{"help", "version", "json", "quiet", "no-color", "debug"},
	Alias: map[string]string{"h": "help", "v": "version", "q": "quiet"},
}

var tokensOpts = Options{
	String: []string{"category", "family"},
	Alias:  map[string]string{"c": "category", "f": "family"},
}

func TestBooleanFlags(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		set  []string // flags expected truthy
		pos  []string
	}{
		{"none", nil, nil, nil},
		{"long", []string{"--json"}, []string{"json"}, nil},
		{"alias", []string{"-q"}, []string{"quiet"}, nil},
		{"anywhere", []string{"--json", "list", "-q"}, []string{"json", "quiet"}, []string{"list"}},
		{"no-color is a plain flag", []string{"--no-color", "list"}, []string{"no-color"}, []string{"list"}},
		{"grouped short flags", []string{"-qv"}, []string{"quiet", "version"}, nil},
		{"= false", []string{"--json=false", "list"}, nil, []string{"list"}},
		{"= true", []string{"--json=true"}, []string{"json"}, nil},
		{"explicit false value", []string{"--json", "false", "list"}, nil, []string{"list"}},
		{"explicit true value", []string{"list", "--json", "true"}, []string{"json"}, []string{"list"}},
		{"boolean flags never take a word", []string{"--json", "list"}, []string{"json"}, []string{"list"}},
		{"short = value is not boolean-parsed", []string{"-q=false"}, []string{"quiet"}, nil},
		{"after double dash everything is positional", []string{"list", "--", "--json", "-q"}, nil, []string{"list", "--json", "-q"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Parse(tt.argv, mainOpts)
			for _, name := range []string{"help", "version", "json", "quiet", "no-color", "debug"} {
				if got, want := r.Truthy(name), slices.Contains(tt.set, name); got != want {
					t.Errorf("Truthy(%q) = %v, want %v", name, got, want)
				}
			}
			if !slices.Equal(r.Positionals, tt.pos) {
				t.Errorf("Positionals = %q, want %q", r.Positionals, tt.pos)
			}
		})
	}
}

func TestAliasesSetEveryName(t *testing.T) {
	r := Parse([]string{"-h"}, mainOpts)
	if !r.Truthy("help") || !r.Truthy("h") {
		t.Error("alias and name should both be set")
	}
}

func TestStringFlags(t *testing.T) {
	tests := []struct {
		name        string
		argv        []string
		category    *string
		family      *string
		positionals []string
	}{
		{name: "long with space", argv: []string{"--category", "color"}, category: ptr("color")},
		{name: "long with =", argv: []string{"--category=color"}, category: ptr("color")},
		{name: "short with space", argv: []string{"-c", "color"}, category: ptr("color")},
		{name: "short with =", argv: []string{"-c=color"}, category: ptr("color")},
		{name: "both", argv: []string{"-c", "color", "-f", "red"}, category: ptr("color"), family: ptr("red")},
		{name: "missing value at end", argv: []string{"--category"}, category: ptr("")},
		{name: "empty = value", argv: []string{"--category="}, category: ptr("")},
		{name: "followed by a flag", argv: []string{"--category", "--json"}, category: ptr("")},
		{name: "value that looks like a short flag is not consumed", argv: []string{"-c", "-x"}, category: ptr("")},
		{name: "last one wins", argv: []string{"-c", "a", "--category", "b"}, category: ptr("b")},
		{name: "value may be 'true'", argv: []string{"--category", "true"}, category: ptr("true")},
		{name: "unset", argv: []string{"list"}, positionals: []string{"list"}},
		{name: "unknown flag swallows the next word", argv: []string{"--foo", "list", "-c", "x"}, category: ptr("x")},
		{name: "unknown flag then a flag", argv: []string{"--foo", "-c", "x"}, category: ptr("x")},
		{name: "empty argument is a positional", argv: []string{"--category", "", "x"}, category: ptr(""), positionals: []string{"", "x"}},
		{name: "positionals are kept in order", argv: []string{"a", "-c", "x", "b"}, category: ptr("x"), positionals: []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Parse(tt.argv, tokensOpts)
			check := func(name string, want *string) {
				got, ok := r.String(name)
				switch {
				case want == nil && ok:
					t.Errorf("String(%q) = %q, want unset", name, got)
				case want != nil && (!ok || got != *want):
					t.Errorf("String(%q) = %q, %v; want %q", name, got, ok, *want)
				}
			}
			check("category", tt.category)
			check("family", tt.family)
			if !slices.Equal(r.Positionals, tt.positionals) {
				t.Errorf("Positionals = %q, want %q", r.Positionals, tt.positionals)
			}
		})
	}
}

func TestStringAliasesShareValue(t *testing.T) {
	r := Parse([]string{"-c", "color"}, tokensOpts)
	long, _ := r.String("category")
	short, _ := r.String("c")
	if long != "color" || short != "color" {
		t.Errorf("category=%q c=%q", long, short)
	}
}

func ptr(s string) *string { return &s }
