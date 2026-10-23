// Package errs defines the user-facing errors mint reports, and the
// "did you mean?" helper used for mistyped subcommands.
package errs

// Error is a failure with a message meant for the user and an optional hint
// about how to fix it.
type Error struct {
	// Kind names the failure class (e.g. "FetchError"); shown by --debug.
	Kind       string
	Message    string
	Suggestion string
	// Err is the underlying cause, if any; shown by --debug.
	Err error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.Err }

// Fetch is returned when the token source cannot be downloaded or understood.
func Fetch(message, suggestion string, cause error) *Error {
	return &Error{Kind: "FetchError", Message: message, Suggestion: suggestion, Err: cause}
}

// Validation is returned when user-supplied input is invalid.
func Validation(message, suggestion string) *Error {
	return &Error{Kind: "ValidationError", Message: message, Suggestion: suggestion}
}

// FileWrite is returned when writing to the filesystem fails.
func FileWrite(message, suggestion string, cause error) *Error {
	return &Error{Kind: "FileWriteError", Message: message, Suggestion: suggestion, Err: cause}
}

// levenshtein computes the edit distance between a and b.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

// Suggest returns the closest entry of valid to input if its edit distance is
// at most 2. Ties go to the earlier entry.
func Suggest(input string, valid []string) (string, bool) {
	best, bestDistance := "", -1
	for _, candidate := range valid {
		if d := levenshtein(input, candidate); bestDistance < 0 || d < bestDistance {
			best, bestDistance = candidate, d
		}
	}
	return best, bestDistance >= 0 && bestDistance <= 2
}
