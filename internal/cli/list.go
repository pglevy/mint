package cli

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/pglevy/mint/internal/args"
	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/jsnum"
	"github.com/pglevy/mint/internal/jsonv"
	"github.com/pglevy/mint/internal/output"
	"github.com/pglevy/mint/internal/tokens"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// listCommand displays all tokens in human-readable or JSON format,
// optionally filtered by category.
func listCommand(s *session, argv []string) error {
	parsed := args.Parse(argv, args.Options{
		String: []string{"category"},
		Alias:  map[string]string{"c": "category"},
	})

	category, hasCategory := parsed.String("category")

	if hasCategory {
		if _, ok := tokens.ParseCategory(category); !ok {
			return errs.Validation(
				fmt.Sprintf(`Unknown category "%s".`, category),
				"Valid categories: "+strings.Join(tokens.CategoryNames(), ", "),
			)
		}
	}

	// Prompt for a category when missing, unless output is piped/redirected
	// (--json or a non-TTY stream).
	if !hasCategory {
		switch {
		case s.env.interactive() && !s.globals.json:
			selected, ok, err := s.env.Select("Select a category:", tokens.CategoryNames())
			if err != nil {
				return err
			}
			if !ok {
				return exitError(1)
			}
			category, hasCategory = selected, true
		case s.globals.json:
			// JSON mode without a category outputs all tokens.
		default:
			s.out.Stderr("Error: --category is required when stdin is not a TTY")
			return exitError(1)
		}
	}

	var filter tokens.Category
	if hasCategory {
		filter = tokens.Category(category)
	}

	doc, err := s.env.Fetch()
	if err != nil {
		return err
	}
	store := tokens.NewStore()
	store.Load(doc)

	entries := store.List(filter)

	if s.globals.json {
		list := make([]jsonv.Value, len(entries))
		for i, e := range entries {
			list[i] = e.JSON()
		}
		s.out.JSON(list)
		return nil
	}

	// Group by category, each with a colored header and aligned columns.
	for len(entries) > 0 {
		cat := entries[0].Category
		n := 1
		for n < len(entries) && entries[n].Category == cat {
			n++
		}
		s.printCategory(cat, entries[:n])
		entries = entries[n:]
	}
	return nil
}

func (s *session) printCategory(cat tokens.Category, entries []tokens.Entry) {
	if !s.globals.quiet {
		header := s.color.Category(fmt.Sprintf("%s (%d tokens)", cat, len(entries)), cat)
		s.out.Stdout("")
		s.out.Stdout("  " + header)
	}

	// Sort: primary by family (alpha), secondary by numeric value (low to high).
	// Entries with no numeric value keep their original order, last.
	collator := collate.New(language.Und)
	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(a, b tokens.Entry) int {
		famA, famB := family(stripPrefix(a.Path)), family(stripPrefix(b.Path))
		if c := collator.CompareString(famA, famB); c != 0 {
			return c
		}
		// Infinity − Infinity is NaN, which compares as equal.
		diff := numericValue(a.Value) - numericValue(b.Value)
		switch {
		case diff < 0:
			return -1
		case diff > 0:
			return 1
		}
		return 0
	})

	paths := make([]string, len(sorted))
	maxPath := 0
	for i, e := range sorted {
		paths[i] = stripPrefix(e.Path)
		maxPath = max(maxPath, utf8.RuneCountInString(paths[i]))
	}

	for i, e := range sorted {
		// Blank line between families.
		if i > 0 && family(paths[i-1]) != family(paths[i]) {
			s.out.Stdout("")
		}
		s.out.Stdout(fmt.Sprintf("    %-*s  %s", maxPath, paths[i], output.FormatValue(e.Value)))
	}
}

// stripPrefix drops the category from a path ("spacing.radius.sm" → "radius.sm").
func stripPrefix(path string) string {
	_, rest, found := strings.Cut(path, ".")
	if found {
		return rest
	}
	return path
}

// family returns the first segment of a stripped path.
func family(path string) string {
	head, _, _ := strings.Cut(path, ".")
	return head
}

// numericValue extracts a sortable number from a token value: numbers as is,
// strings by their leading number ("16px" → 16), composite tokens by their
// "value" key. Anything else sorts last.
func numericValue(v jsonv.Value) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case string:
		if n := jsnum.ParseFloat(x); !math.IsNaN(n) {
			return n
		}
	case *jsonv.Object:
		inner, _ := x.Get("value")
		switch y := inner.(type) {
		case float64:
			return y
		case string:
			if n := jsnum.ParseFloat(y); !math.IsNaN(n) {
				return n
			}
		}
	}
	return math.Inf(1)
}
