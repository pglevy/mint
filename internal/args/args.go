// Package args parses command-line arguments with the same rules as Deno's
// @std/cli parseArgs (v1.0.29), which the original mint CLI used.
//
// mint's behaviour depends on its leniency: global flags such as --json may
// appear before or after the subcommand, unknown flags are accepted rather
// than rejected, and a flag followed by a bare word takes it as its value.
// Replicating those rules (instead of adopting a stricter parser) keeps every
// existing invocation working the same way.
//
// Only the subset of options mint uses is supported: boolean and string flags
// with aliases. Two minor simplifications are deliberate: numeric-looking
// values stay strings (so `-q=0` is truthy), and dotted names such as --a.b
// are plain keys rather than nested objects.
package args

import (
	"regexp"
	"strings"
)

// Options declares which flags are boolean or string-valued, and their
// aliases, e.g. {"h": "help"}.
type Options struct {
	Bool   []string
	String []string
	Alias  map[string]string
}

// Result holds the parsed flags and positional arguments.
type Result struct {
	// Positionals are the arguments that are not flags, in order.
	Positionals []string
	values      map[string]any // string or bool
}

// Truthy reports whether the flag is set to a truthy value (JavaScript's
// Boolean(parsed[name])): true, or a non-empty string.
func (r Result) Truthy(name string) bool {
	switch v := r.values[name].(type) {
	case bool:
		return v
	case string:
		return v != ""
	}
	return false
}

// String returns the flag's string value and whether it was given at all.
// A string flag with no value (`--category` at the end) yields "".
func (r Result) String(name string) (string, bool) {
	switch v := r.values[name].(type) {
	case string:
		return v, true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	}
	return "", false
}

var (
	flagRegexp    = regexp.MustCompile(`(?s)^-(?:(?P<doubleDash>-)(?P<negated>no-)?)?(?P<key>.+?)(?:=(?P<value>.*))?$`)
	letterRegexp  = regexp.MustCompile(`[A-Za-z]`)
	numberRegexp  = regexp.MustCompile(`-?\d+(\.\d*)?(e-?\d+)?$`)
	hyphenRegexp  = regexp.MustCompile(`^(-|--)[^-]`)
	valueRegexp   = regexp.MustCompile(`=(.+)`)
	specialRegexp = regexp.MustCompile(`\W`)

	idxDoubleDash = flagRegexp.SubexpIndex("doubleDash")
	idxNegated    = flagRegexp.SubexpIndex("negated")
	idxKey        = flagRegexp.SubexpIndex("key")
	idxValue      = flagRegexp.SubexpIndex("value")
)

type parser struct {
	bools   map[string]bool
	strs    map[string]bool
	aliases map[string][]string
	values  map[string]any
}

// Parse parses argv according to opts.
func Parse(argv []string, opts Options) Result {
	p := &parser{
		bools:   map[string]bool{},
		strs:    map[string]bool{},
		aliases: map[string][]string{},
		values:  map[string]any{},
	}
	for name, target := range opts.Alias {
		p.aliases[name] = []string{target}
	}
	for name, target := range opts.Alias {
		p.aliases[target] = append(p.aliases[target], name)
	}
	for _, key := range opts.Bool {
		p.bools[key] = true
		for _, al := range p.aliases[key] {
			p.bools[al] = true
		}
	}
	for _, key := range opts.String {
		p.strs[key] = true
		for _, al := range p.aliases[key] {
			p.strs[al] = true
		}
	}

	// Everything after the first "--" is positional.
	var notFlags []string
	for i, a := range argv {
		if a == "--" {
			notFlags = argv[i+1:]
			argv = argv[:i]
			break
		}
	}

	var positionals []string
argsLoop:
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		m := flagRegexp.FindStringSubmatchIndex(arg)
		if m == nil {
			positionals = append(positionals, arg)
			continue
		}
		group := func(idx int) (string, bool) {
			if m[2*idx] < 0 {
				return "", false
			}
			return arg[m[2*idx]:m[2*idx+1]], true
		}

		if _, isDouble := group(idxDoubleDash); isDouble {
			key, _ := group(idxKey)

			if value, ok := group(idxValue); ok {
				if p.bools[key] {
					p.set(key, value != "false")
				} else {
					p.set(key, value)
				}
				continue
			}

			// "--no-foo" is only a negation for names declared negatable, and
			// mint declares none, so it is the flag literally named "no-foo".
			if _, negated := group(idxNegated); negated {
				key = "no-" + key
			}

			if next := argAt(argv, i+1); next != "" {
				if !p.bools[key] && !strings.HasPrefix(next, "-") && !p.aliasIsBool(key) {
					p.set(key, next)
					i++
					continue
				}
				if next == "true" || next == "false" {
					p.set(key, next == "true")
					i++
					continue
				}
			}
			p.set(key, p.defaultValue(key))
			continue
		}

		// Short flags: "-q", "-qv", "-c=color", "-c color".
		chars := []rune(arg)
		letters := chars[1 : len(chars)-1]
		for j, letter := range letters {
			next := string(chars[j+2:])

			if next == "-" {
				p.set(string(letter), next)
				continue
			}
			if next == "=" {
				p.set(string(letter), "")
				continue argsLoop
			}
			if letterRegexp.MatchString(string(letter)) {
				if g := valueRegexp.FindStringSubmatch(next); g != nil {
					p.set(string(letter), g[1])
					continue argsLoop
				}
				if numberRegexp.MatchString(next) {
					p.set(string(letter), next)
					continue argsLoop
				}
			}
			if j+1 < len(letters) && specialRegexp.MatchString(string(letters[j+1])) {
				p.set(string(letter), next)
				continue argsLoop
			}
			p.set(string(letter), p.defaultValue(string(letter)))
		}

		key := string(chars[len(chars)-1])
		if key == "-" {
			continue
		}
		if next := argAt(argv, i+1); next != "" {
			if !hyphenRegexp.MatchString(next) && !p.bools[key] && !p.aliasIsBool(key) {
				p.set(key, next)
				i++
				continue
			}
			if next == "true" || next == "false" {
				p.set(key, next == "true")
				i++
				continue
			}
		}
		p.set(key, p.defaultValue(key))
	}

	return Result{
		Positionals: append(positionals, notFlags...),
		values:      p.values,
	}
}

// argAt returns argv[i], or "" past the end. (parseArgs treats an empty
// argument like a missing one when looking for a flag's value.)
func argAt(argv []string, i int) string {
	if i < len(argv) {
		return argv[i]
	}
	return ""
}

// set records value under key and all of its aliases.
func (p *parser) set(key string, value any) {
	p.values[key] = value
	for _, al := range p.aliases[key] {
		p.values[al] = value
	}
}

// defaultValue is what a flag with no explicit value gets.
func (p *parser) defaultValue(key string) any {
	if p.strs[key] {
		return ""
	}
	return true
}

// aliasIsBool reports whether key has an alias that is declared boolean.
func (p *parser) aliasIsBool(key string) bool {
	for _, al := range p.aliases[key] {
		if p.bools[al] {
			return true
		}
	}
	return false
}
