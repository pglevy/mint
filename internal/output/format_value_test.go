package output

import (
	"strings"
	"testing"

	"github.com/pglevy/mint/internal/jsonv"
)

func parse(t *testing.T, s string) jsonv.Value {
	t.Helper()
	v, err := jsonv.Parse([]byte(s))
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{"string passthrough", `"#ff0000"`, "#ff0000"},
		{"string with unit", `"16px"`, "16px"},
		{"empty string", `""`, ""},
		{"integer", `16`, "16"},
		{"zero", `0`, "0"},
		{"decimal", `3.14`, "3.14"},
		{"object with sorted keys", `{"fontSize":"16px","color":"#000"}`, "color: #000, fontSize: 16px"},
		{"nested objects are JSON", `{"shadow":{"x":0,"y":4}}`, `shadow: {"x":0,"y":4}`},
		{"composite typography token", `{"fontSize":"1.5rem","lineHeight":"2","fontWeight":"700"}`,
			"fontSize: 1.5rem, fontWeight: 700, lineHeight: 2"},
		{"boolean and null nested values", `{"active":true,"removed":false,"ref":null}`,
			"active: true, ref: null, removed: false"},
		{"nested array is JSON", `{"tags":["a","b"]}`, `tags: ["a","b"]`},
		{"nested number", `{"lineHeight":1.25}`, "lineHeight: 1.25"},
		{"value and unit combine", `{"value":0.5,"unit":"rem"}`, "0.5rem"},
		{"value and unit combine (zero)", `{"value":0,"unit":"px"}`, "0px"},
		{"value and unit with a third key stays an object", `{"value":1,"unit":"px","extra":2}`,
			"extra: 2, unit: px, value: 1"},
		{"value without unit stays an object", `{"value":1}`, "value: 1"},
		{"gradient stops", `[{"color":"#fff","position":0},{"color":"#000","position":0.575},{"color":"red","position":1}]`,
			"0% #fff · 57% #000 · 100% red"},
		{"gradient stop positions round half up", `[{"color":"a","position":0.005},{"color":"b","position":0.125},{"color":"c","position":0.995}]`,
			"1% a · 13% b · 100% c"},
		{"gradient stop with numeric string position", `[{"color":"a","position":"0.5"}]`, "50% a"},
		{"array of strings is JSON", `["Open Sans","system-ui"]`, `["Open Sans","system-ui"]`},
		{"empty array is JSON", `[]`, `[]`},
		{"array that is not stops is JSON", `[1,2,3]`, `[1,2,3]`},
		// The Deno implementation printed "50% #fff · NaN% undefined" here.
		{"malformed stops fall back to JSON", `[{"color":"#fff","position":0.5},7]`, `[{"color":"#fff","position":0.5},7]`},
		{"empty object", `{}`, ""},
		{"boolean", `true`, "true"},
		{"null", `null`, "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(parse(t, tt.json))
			if got != tt.want {
				t.Errorf("FormatValue(%s) = %q, want %q", tt.json, got, tt.want)
			}
			if strings.Contains(got, "[object Object]") {
				t.Errorf("output contains [object Object]: %q", got)
			}
		})
	}
}
