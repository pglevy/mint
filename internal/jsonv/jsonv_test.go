package jsonv

import (
	"math"
	"strings"
	"testing"
)

func mustParse(t *testing.T, s string) Value {
	t.Helper()
	v, err := Parse([]byte(s))
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return v
}

func TestKeyOrderFollowsJavaScript(t *testing.T) {
	// Array-index keys come first in ascending numeric order; the rest keep
	// insertion order. "01", "-1", "1.5" and 2³²−1 are not array indices.
	in := `{"b":1,"10":1,"a":1,"2":1,"01":1,"-1":1,"1.5":1,"4294967295":1,"4294967294":1,"0":1,"c":1}`
	want := []string{"0", "2", "10", "4294967294", "b", "a", "01", "-1", "1.5", "4294967295", "c"}
	got := mustParse(t, in).(*Object).Keys()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("keys = %v, want %v", got, want)
	}
}

func TestDuplicateKeysKeepFirstPositionAndLastValue(t *testing.T) {
	obj := mustParse(t, `{"a":1,"b":2,"a":3}`).(*Object)
	if got := Marshal(obj, ""); got != `{"a":3,"b":2}` {
		t.Errorf("got %s", got)
	}
}

func TestMarshalLayout(t *testing.T) {
	v := mustParse(t, `{"a":[1,2,{"b":null}],"c":{},"d":[],"e":"x"}`)
	want := `{
  "a": [
    1,
    2,
    {
      "b": null
    }
  ],
  "c": {},
  "d": [],
  "e": "x"
}`
	if got := Marshal(v, "  "); got != want {
		t.Errorf("indented:\n%s\nwant:\n%s", got, want)
	}
	if got, want := Marshal(v, ""), `{"a":[1,2,{"b":null}],"c":{},"d":[],"e":"x"}`; got != want {
		t.Errorf("compact: %s, want %s", got, want)
	}
}

func TestMarshalStrings(t *testing.T) {
	// JSON.stringify escapes quotes, backslashes and control characters, but
	// not HTML-sensitive characters, DEL, U+2028 or non-ASCII text.
	in := "quote \" back \\ tab \t nl \n cr \r bs \b ff \f bell \a del \x7f <b>&</b> é 日本 😀 \u2028"
	want := `"quote \" back \\ tab \t nl \n cr \r bs \b ff \f bell \u0007 del ` + "\x7f" + ` <b>&</b> é 日本 😀 ` + "\u2028" + `"`
	if got := Marshal(in, ""); got != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestMarshalNumbers(t *testing.T) {
	tests := map[string]string{
		"1.0": "1", "100": "100", "1e2": "100", "0.10": "0.1", "-0": "0", "1e21": "1e+21",
		"1e-7": "1e-7", "123456789012345680000": "123456789012345680000", "0.000001": "0.000001",
	}
	for in, want := range tests {
		if got := Marshal(mustParse(t, in), ""); got != want {
			t.Errorf("Marshal(%s) = %s, want %s", in, got, want)
		}
	}
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := Marshal(f, ""); got != "null" {
			t.Errorf("Marshal(%v) = %s, want null", f, got)
		}
	}
	// A literal beyond float64 range becomes Infinity, which stringifies as null.
	if got := Marshal(mustParse(t, "1e999"), ""); got != "null" {
		t.Errorf("Marshal(1e999) = %s, want null", got)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "{", `{"a":}`, `[1,]`, `{"a":1} x`, `<!DOCTYPE html>`, `{"a":1}{"b":2}`, `nul`} {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", in)
		}
	}
}

func TestParseAcceptsScalarsAndBOM(t *testing.T) {
	if v := mustParse(t, "\ufeff{\"a\":1}"); v.(*Object).Len() != 1 {
		t.Errorf("BOM not skipped: %v", v)
	}
	if v := mustParse(t, ` "x" `); v != "x" {
		t.Errorf("got %v", v)
	}
	if v := mustParse(t, "null"); v != nil {
		t.Errorf("got %v", v)
	}
}

func TestCloneIsDeep(t *testing.T) {
	orig := mustParse(t, `{"a":{"b":[1,{"c":2}]}}`).(*Object)
	clone := Clone(orig).(*Object)
	inner, _ := clone.Get("a")
	inner.(*Object).Set("b", "changed")
	if got := Marshal(orig, ""); got != `{"a":{"b":[1,{"c":2}]}}` {
		t.Errorf("original modified: %s", got)
	}
}

func TestSetKeepsPositionOnOverwrite(t *testing.T) {
	o := NewObject()
	o.Set("z", 1.0)
	o.Set("5", 1.0)
	o.Set("a", 1.0)
	o.Set("z", 2.0)
	o.Set("1", 1.0)
	if got := Marshal(o, ""); got != `{"1":1,"5":1,"z":2,"a":1}` {
		t.Errorf("got %s", got)
	}
}
