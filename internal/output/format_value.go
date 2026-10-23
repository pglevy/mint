package output

import (
	"math"
	"sort"
	"strings"

	"github.com/pglevy/mint/internal/jsnum"
	"github.com/pglevy/mint/internal/jsonv"
)

// FormatValue formats a token $value as a human-readable string.
//
//   - strings pass through unchanged and numbers are printed as in JavaScript
//   - gradient stops [{color, position}, …] become "0% #fff · 57% #000"
//   - an object with exactly "value" and "unit" becomes "0.5rem"
//   - any other object becomes sorted "key: value" pairs joined by ", ";
//     nested objects and arrays are rendered as compact JSON
func FormatValue(v jsonv.Value) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return jsnum.Format(x)
	case bool, nil:
		return jsonv.Marshal(x, "")
	case []jsonv.Value:
		if stops, ok := gradientStops(x); ok {
			return stops
		}
		return jsonv.Marshal(x, "")
	case *jsonv.Object:
		return formatObject(x)
	}
	return jsonv.Marshal(v, "")
}

// gradientStops formats a non-empty array whose elements all look like
// {color, position}: "0% #fff · 57% #000". A position may be a number or a
// numeric string. Arrays that don't fit fall back to compact JSON.
func gradientStops(arr []jsonv.Value) (string, bool) {
	if len(arr) == 0 {
		return "", false
	}
	parts := make([]string, len(arr))
	for i, e := range arr {
		stop, ok := e.(*jsonv.Object)
		if !ok {
			return "", false
		}
		color, _ := stop.Get("color")
		position, _ := stop.Get("position")
		c, isColor := scalar(color)
		p, isPosition := number(position)
		if !isColor || !isPosition {
			return "", false
		}
		parts[i] = jsnum.Format(jsnum.Round(p*100)) + "% " + c
	}
	return strings.Join(parts, " · "), true
}

// number converts a number or numeric string, as JavaScript's arithmetic would.
func number(v jsonv.Value) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		n := jsnum.ToNumber(x)
		return n, !math.IsNaN(n)
	}
	return 0, false
}

func formatObject(obj *jsonv.Object) string {
	if obj.Len() == 2 && obj.Has("value") && obj.Has("unit") {
		value, _ := obj.Get("value")
		unit, _ := obj.Get("unit")
		if v, ok := scalar(value); ok {
			if u, ok := scalar(unit); ok {
				return v + u
			}
		}
	}

	keys := append([]string(nil), obj.Keys()...)
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, key := range keys {
		value, _ := obj.Get(key)
		parts[i] = key + ": " + formatNested(value)
	}
	return strings.Join(parts, ", ")
}

// scalar renders primitives (strings, numbers, booleans, null) as JavaScript's
// string conversion would. Objects and arrays are not scalars.
func scalar(v jsonv.Value) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		return jsnum.Format(x), true
	case bool, nil:
		return jsonv.Marshal(x, ""), true
	}
	return "", false
}

// formatNested formats a value one level inside an object: primitives print
// bare, objects and arrays as compact JSON.
func formatNested(v jsonv.Value) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return jsnum.Format(x)
	}
	return jsonv.Marshal(v, "")
}
