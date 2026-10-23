package cli

import (
	"fmt"
	"strings"

	"github.com/pglevy/mint/internal/args"
	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/jsonv"
	"github.com/pglevy/mint/internal/tokens"
)

// tokensCommand outputs raw DTCG token data as formatted JSON, optionally
// filtered by --category and --family. It always outputs JSON: this is the
// "raw data" command.
func tokensCommand(s *session, argv []string) error {
	parsed := args.Parse(argv, args.Options{
		String: []string{"category", "family"},
		Alias:  map[string]string{"c": "category", "f": "family"},
	})

	category, hasCategory := parsed.String("category")
	family, hasFamily := parsed.String("family")

	if hasCategory {
		if _, ok := tokens.ParseCategory(category); !ok {
			return errs.Validation(
				fmt.Sprintf(`Unknown category "%s". Valid categories: %s`,
					category, strings.Join(tokens.CategoryNames(), ", ")),
				"",
			)
		}
	}

	// Prompt for a category when missing.
	if !hasCategory {
		switch {
		case s.env.interactive():
			selected, ok, err := s.env.Select("Select a category:", tokens.CategoryNames())
			if err != nil {
				return err
			}
			if !ok {
				return exitError(1)
			}
			category, hasCategory = selected, true
		case !s.env.StdoutIsTTY:
			// Output is piped/redirected: dump all categories.
		default:
			s.out.Stderr("Error: --category is required when stdin is not a TTY")
			return exitError(1)
		}
	}

	doc, err := s.env.Fetch()
	if err != nil {
		return err
	}
	store := tokens.NewStore()
	store.Load(doc)

	// Prompt for a family when the category is known but the family is not.
	// Without a terminal the family is simply optional.
	if !hasFamily && hasCategory {
		families := store.Families(tokens.Category(category))
		if len(families) > 0 && s.env.interactive() {
			selected, ok, err := s.env.Select(fmt.Sprintf(`Select a family in "%s":`, category), families)
			if err != nil {
				return err
			}
			if !ok {
				return exitError(1)
			}
			family, hasFamily = selected, true
		}
	}

	var result jsonv.Value
	switch {
	case hasCategory && family != "":
		result = store.ByFamily(tokens.Category(category), family)
	case hasCategory:
		result = store.ByCategory(tokens.Category(category))
	default:
		result = store.All()
	}

	s.out.Stdout(jsonv.Marshal(result, "  "))
	return nil
}
