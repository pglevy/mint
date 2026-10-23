package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/pglevy/mint/internal/args"
	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/tokens"
)

// exportCommand generates DTCG JSON and writes it to stdout or, with
// --output, to a file.
func exportCommand(s *session, argv []string) error {
	parsed := args.Parse(argv, args.Options{
		String: []string{"category", "output"},
		Alias:  map[string]string{"c": "category", "o": "output"},
	})

	category, hasCategory := parsed.String("category")
	outputPath, _ := parsed.String("output")

	if hasCategory {
		if _, ok := tokens.ParseCategory(category); !ok {
			return errs.Validation(
				fmt.Sprintf(`Unknown category "%s".`, category),
				"Valid categories: "+strings.Join(tokens.CategoryNames(), ", "),
			)
		}
	}

	doc, err := s.env.Fetch()
	if err != nil {
		return err
	}
	store := tokens.NewStore()
	store.Load(doc)

	dtcg := store.JSON(tokens.Category(category))

	if outputPath == "" {
		s.out.Stdout(dtcg)
		return nil
	}

	if err := os.WriteFile(outputPath, []byte(dtcg+"\n"), 0o666); err != nil {
		return errs.FileWrite(
			fmt.Sprintf(`Cannot write to "%s".`, outputPath),
			"Check that the directory exists and you have write permissions.",
			err,
		)
	}
	// Confirm on stderr unless --quiet.
	s.out.Stderr("Wrote tokens to " + outputPath)
	return nil
}
