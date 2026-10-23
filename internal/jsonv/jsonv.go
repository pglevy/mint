// Package jsonv models JSON documents the way JavaScript does, so that mint
// prints token data exactly as the Deno implementation did.
//
// Go's map[string]any loses key order and encoding/json sorts keys, which
// would scramble token scales such as "50", "100", "200". JavaScript objects
// instead keep insertion order, except that integer-like keys ("50") always
// come first in ascending numeric order. Object reproduces that rule, and
// Marshal reproduces JSON.stringify's layout, escaping and number formatting.
//
// A Value is one of: nil, bool, float64, string, []Value or *Object.
package jsonv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pglevy/mint/internal/jsnum"
)

// Value is a decoded JSON value: nil, bool, float64, string, []Value or *Object.
type Value = any

// Object is a JSON object that iterates keys in JavaScript property order.
type Object struct {
	keys    []string
	vals    map[string]Value
	indexed int // number of leading keys that are array indices
}

// NewObject returns an empty Object.
func NewObject() *Object {
	return &Object{vals: map[string]Value{}}
}

// Set stores v under key. As in JavaScript, overwriting an existing key keeps
// its position.
func (o *Object) Set(key string, v Value) {
	if _, ok := o.vals[key]; ok {
		o.vals[key] = v
		return
	}
	o.vals[key] = v
	n, isIndex := arrayIndex(key)
	if !isIndex {
		o.keys = append(o.keys, key)
		return
	}
	// Array-index keys form a numerically sorted prefix.
	pos := sort.Search(o.indexed, func(i int) bool {
		m, _ := arrayIndex(o.keys[i])
		return m > n
	})
	o.keys = append(o.keys, "")
	copy(o.keys[pos+1:], o.keys[pos:])
	o.keys[pos] = key
	o.indexed++
}

// Get returns the value stored under key.
func (o *Object) Get(key string) (Value, bool) {
	v, ok := o.vals[key]
	return v, ok
}

// Has reports whether key is present.
func (o *Object) Has(key string) bool {
	_, ok := o.vals[key]
	return ok
}

// Keys returns the keys in JavaScript property order. The slice must not be
// modified.
func (o *Object) Keys() []string { return o.keys }

// Len returns the number of keys.
func (o *Object) Len() int { return len(o.keys) }

// arrayIndex reports whether key is a canonical array index (0 … 2³²−2), the
// keys JavaScript engines order numerically ahead of all others.
func arrayIndex(key string) (uint32, bool) {
	if key == "" || (len(key) > 1 && key[0] == '0') || len(key) > 10 {
		return 0, false
	}
	for i := 0; i < len(key); i++ {
		if key[i] < '0' || key[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseUint(key, 10, 64)
	if err != nil || n > math.MaxUint32-1 {
		return 0, false
	}
	return uint32(n), true
}

// Clone returns a deep copy of v.
func Clone(v Value) Value {
	switch x := v.(type) {
	case *Object:
		c := NewObject()
		for _, k := range x.keys {
			c.Set(k, Clone(x.vals[k]))
		}
		return c
	case []Value:
		c := make([]Value, len(x))
		for i, e := range x {
			c[i] = Clone(e)
		}
		return c
	default:
		return v
	}
}

// Parse decodes a JSON document, like JSON.parse. Duplicate keys keep the
// first position and the last value.
func Parse(data []byte) (Value, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")) // fetch().json() strips a BOM
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decode(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data after top-level value")
	}
	return v, nil
}

func decode(dec *json.Decoder) (Value, error) {
	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		if t == '[' {
			arr := []Value{}
			for dec.More() {
				e, err := decode(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, e)
			}
			_, err := dec.Token() // ]
			return arr, err
		}
		obj := NewObject()
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			val, err := decode(dec)
			if err != nil {
				return nil, err
			}
			obj.Set(keyTok.(string), val)
		}
		_, err := dec.Token() // }
		return obj, err
	case json.Number:
		// Out-of-range literals (1e999) become ±Infinity, as in JavaScript.
		f, err := strconv.ParseFloat(string(t), 64)
		if err != nil && !errors.Is(err, strconv.ErrRange) {
			return nil, err
		}
		return f, nil
	default: // string, bool, nil
		return tok, nil
	}
}

// Marshal formats v like JSON.stringify(v, null, indent) where indent is the
// per-level indentation ("" for compact output).
func Marshal(v Value, indent string) string {
	var sb strings.Builder
	write(&sb, v, indent, "")
	return sb.String()
}

func write(sb *strings.Builder, v Value, indent, cur string) {
	switch x := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		sb.WriteString(strconv.FormatBool(x))
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			sb.WriteString("null")
		} else {
			sb.WriteString(jsnum.Format(x))
		}
	case string:
		writeString(sb, x)
	case []Value:
		if len(x) == 0 {
			sb.WriteString("[]")
			return
		}
		sb.WriteByte('[')
		inner := cur + indent
		for i, e := range x {
			if i > 0 {
				sb.WriteByte(',')
			}
			newline(sb, indent, inner)
			write(sb, e, indent, inner)
		}
		newline(sb, indent, cur)
		sb.WriteByte(']')
	case *Object:
		if x.Len() == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteByte('{')
		inner := cur + indent
		for i, k := range x.keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			newline(sb, indent, inner)
			writeString(sb, k)
			sb.WriteByte(':')
			if indent != "" {
				sb.WriteByte(' ')
			}
			write(sb, x.vals[k], indent, inner)
		}
		newline(sb, indent, cur)
		sb.WriteByte('}')
	default:
		panic(fmt.Sprintf("jsonv: unsupported value type %T", v))
	}
}

func newline(sb *strings.Builder, indent, cur string) {
	if indent != "" {
		sb.WriteByte('\n')
		sb.WriteString(cur)
	}
}

const hex = "0123456789abcdef"

// writeString quotes s the way JSON.stringify does: only the quote, the
// backslash and control characters are escaped (no HTML or U+2028 escaping).
func writeString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for i := 0; i < len(s); {
		c := s[i]
		if c >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size == 1 {
				sb.WriteString(`�`)
			} else {
				sb.WriteString(s[i : i+size])
			}
			i += size
			continue
		}
		switch c {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if c < 0x20 {
				sb.WriteString(`\u00`)
				sb.WriteByte(hex[c>>4])
				sb.WriteByte(hex[c&0xf])
			} else {
				sb.WriteByte(c)
			}
		}
		i++
	}
	sb.WriteByte('"')
}
