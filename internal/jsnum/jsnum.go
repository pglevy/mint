// Package jsnum reproduces the few JavaScript number behaviours that mint's
// output depends on: how numbers are printed (String(n)), how leading numeric
// text is parsed (parseFloat), and how they are rounded (Math.round).
//
// The Deno implementation relied on these implicitly. Keeping them explicit
// keeps `mint` output byte-for-byte identical across the port.
package jsnum

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Format returns the same text as JavaScript's String(n).
//
// Go's strconv produces the same shortest round-trip digits, but lays them out
// differently (it prints 1e+21 as "1e+21" but 100000 as "100000", and switches
// to exponent form at different magnitudes), so the layout follows the
// ECMAScript Number::toString algorithm.
func Format(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case f == 0: // also -0
		return "0"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}

	sign := ""
	if f < 0 {
		sign = "-"
		f = -f
	}

	// Shortest digits that round-trip, as d[.ddd]e±XX.
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(f, 'e', -1, 64), "e")
	exp, _ := strconv.Atoi(exponent)
	digits := strings.Replace(mantissa, ".", "", 1)
	k := len(digits) // number of significant digits
	n := exp + 1     // position of the decimal point relative to the digits

	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}

	e := n - 1
	expSign := "+"
	if e < 0 {
		expSign = "-"
		e = -e
	}
	if k == 1 {
		return sign + digits + "e" + expSign + strconv.Itoa(e)
	}
	return sign + digits[:1] + "." + digits[1:] + "e" + expSign + strconv.Itoa(e)
}

var floatPrefix = regexp.MustCompile(`^[+-]?(?:Infinity|(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)`)

// ParseFloat mirrors JavaScript's parseFloat: leading whitespace is skipped,
// the longest numeric prefix is used ("16px" → 16), and NaN is returned when
// there is no numeric prefix ("#fff" → NaN).
func ParseFloat(s string) float64 {
	s = strings.TrimLeftFunc(s, isSpace)
	prefix := floatPrefix.FindString(s)
	if prefix == "" {
		return math.NaN()
	}
	switch prefix {
	case "Infinity", "+Infinity":
		return math.Inf(1)
	case "-Infinity":
		return math.Inf(-1)
	}
	// Only a range error is possible here, and ParseFloat already returns ±Inf
	// (or 0 on underflow) in that case, which is what JavaScript yields.
	f, _ := strconv.ParseFloat(prefix, 64)
	return f
}

var decimalLiteral = regexp.MustCompile(`^[+-]?(?:Infinity|(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)$`)

// ToNumber mirrors Number(string): surrounding whitespace is ignored, an empty
// string is 0, 0x/0o/0b prefixes select a base, and anything that is not
// entirely a number is NaN ("16px" → NaN, unlike ParseFloat).
func ToNumber(s string) float64 {
	s = strings.TrimFunc(s, isSpace)
	if s == "" {
		return 0
	}
	if len(s) > 2 && s[0] == '0' {
		base := 0
		switch s[1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
		if base != 0 {
			digits := s[2:]
			n, ok := new(big.Int).SetString(digits, base)
			if !ok || digits[0] == '+' || digits[0] == '-' {
				return math.NaN()
			}
			f, _ := new(big.Float).SetInt(n).Float64()
			return f
		}
	}
	if !decimalLiteral.MatchString(s) {
		return math.NaN()
	}
	return ParseFloat(s)
}

// Round mirrors Math.round: halves round toward +Infinity.
func Round(f float64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f
	}
	r := math.Floor(f)
	if f-r >= 0.5 {
		r++
	}
	return r
}

// isSpace reports whether r is JavaScript WhiteSpace or a LineTerminator.
// unicode.IsSpace differs (it includes U+0085 and omits U+FEFF).
func isSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', 0xFEFF, 0x2028, 0x2029:
		return true
	}
	return unicode.Is(unicode.Zs, r)
}
