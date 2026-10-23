package output

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pglevy/mint/internal/jsnum"
	"github.com/pglevy/mint/internal/jsonv"
)

// Property-based round-trip test for FormatValue: random objects (1-5 keys
// holding strings, numbers or nested objects) are formatted and parsed back
// with a test-only inverse, and must equal the original.

func randomKey(r *rand.Rand) string {
	n := 3 + r.Intn(6)
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + r.Intn(26))
	}
	return string(b)
}

// randomString starts with a letter, so it never reads back as a number, and
// uses only letters, digits and spaces (no commas, colons or braces).
func randomString(r *rand.Rand) string {
	const first = "abcdefghijklmnopqrstuvwxyz"
	const rest = "abcdefghijklmnopqrstuvwxyz0123456789 "
	n := 2 + r.Intn(11)
	b := []byte{first[r.Intn(len(first))]}
	for i := 1; i < n; i++ {
		b = append(b, rest[r.Intn(len(rest))])
	}
	if s := strings.TrimRight(string(b), " "); s != "" {
		return s
	}
	return "a"
}

// randomNumber returns an integer or a simple two-decimal number.
func randomNumber(r *rand.Rand) float64 {
	if r.Intn(2) == 0 {
		return float64(r.Intn(1999) - 999)
	}
	return float64(r.Intn(20001)-10000) / 100
}

func randomLeaf(r *rand.Rand) jsonv.Value {
	if r.Intn(2) == 0 {
		return randomString(r)
	}
	return randomNumber(r)
}

func randomNested(r *rand.Rand) *jsonv.Object {
	o := jsonv.NewObject()
	for i, n := 0, 1+r.Intn(3); i < n; i++ {
		o.Set(randomKey(r), randomLeaf(r))
	}
	return o
}

func randomObject(r *rand.Rand) *jsonv.Object {
	o := jsonv.NewObject()
	for i, n := 0, 1+r.Intn(5); i < n; i++ {
		switch key := randomKey(r); r.Intn(3) {
		case 0:
			o.Set(key, randomString(r))
		case 1:
			o.Set(key, randomNumber(r))
		default:
			o.Set(key, randomNested(r))
		}
	}
	return o
}

// parseFormatted is the inverse of FormatValue for objects.
func parseFormatted(formatted string) map[string]jsonv.Value {
	result := map[string]jsonv.Value{}
	for _, pair := range splitPairs(formatted) {
		key, raw, ok := strings.Cut(pair, ": ")
		if !ok {
			continue
		}
		result[key] = parseValue(raw)
	}
	return result
}

// splitPairs splits on ", " outside of JSON braces.
func splitPairs(formatted string) []string {
	var pairs []string
	var cur strings.Builder
	depth := 0
	for i := 0; i < len(formatted); i++ {
		switch ch := formatted[i]; {
		case ch == '{':
			depth++
			cur.WriteByte(ch)
		case ch == '}':
			depth--
			cur.WriteByte(ch)
		case ch == ',' && i+1 < len(formatted) && formatted[i+1] == ' ' && depth == 0:
			pairs = append(pairs, cur.String())
			cur.Reset()
			i++ // skip the space
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		pairs = append(pairs, cur.String())
	}
	return pairs
}

func parseValue(raw string) jsonv.Value {
	if n, err := strconv.ParseFloat(raw, 64); err == nil && jsnum.Format(n) == raw {
		return n
	}
	if strings.HasPrefix(raw, "{") {
		if v, err := jsonv.Parse([]byte(raw)); err == nil {
			return v
		}
	}
	return raw
}

// equalIgnoringOrder compares values, treating objects as unordered.
func equalIgnoringOrder(a, b jsonv.Value) bool {
	ao, aIsObj := a.(*jsonv.Object)
	bo, bIsObj := b.(*jsonv.Object)
	if aIsObj || bIsObj {
		if !aIsObj || !bIsObj || ao.Len() != bo.Len() {
			return false
		}
		for _, k := range ao.Keys() {
			av, _ := ao.Get(k)
			bv, ok := bo.Get(k)
			if !ok || !equalIgnoringOrder(av, bv) {
				return false
			}
		}
		return true
	}
	return a == b
}

func TestFormatValueRoundTripProperty(t *testing.T) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	for i := 0; i < 500; i++ {
		original := randomObject(r)
		formatted := FormatValue(original)
		reconstructed := parseFormatted(formatted)

		if len(reconstructed) != original.Len() {
			t.Fatalf("seed %d iteration %d: %d keys round-tripped, want %d\nOriginal:  %s\nFormatted: %q",
				seed, i+1, len(reconstructed), original.Len(), jsonv.Marshal(original, ""), formatted)
		}
		for _, k := range original.Keys() {
			want, _ := original.Get(k)
			if got, ok := reconstructed[k]; !ok || !equalIgnoringOrder(got, want) {
				t.Fatalf("seed %d iteration %d: key %q round-tripped to %s, want %s\nOriginal:  %s\nFormatted: %q",
					seed, i+1, k, describe(got), describe(want), jsonv.Marshal(original, ""), formatted)
			}
		}
	}
}

func describe(v jsonv.Value) string {
	if v == nil {
		return "<missing>"
	}
	return jsonv.Marshal(v, "")
}
