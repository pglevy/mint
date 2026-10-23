package tokens

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/pglevy/mint/internal/jsonv"
)

func parse(t *testing.T, s string) jsonv.Value {
	t.Helper()
	v, err := jsonv.Parse([]byte(s))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return v
}

func fixture(t *testing.T) jsonv.Value {
	t.Helper()
	data, err := os.ReadFile("testdata/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	return parse(t, string(data))
}

func loadFixture(t *testing.T) *Store {
	t.Helper()
	store := NewStore()
	store.Load(fixture(t))
	return store
}

func TestLoadReturnsCountsPerCategory(t *testing.T) {
	counts := NewStore().Load(fixture(t))
	want := Counts{Color: 4, Gradient: 0, Typography: 2, Spacing: 2}
	for c, n := range want {
		if counts[c] != n {
			t.Errorf("counts[%s] = %d, want %d", c, counts[c], n)
		}
	}
}

func TestByCategoryReturnsOnlyThatCategory(t *testing.T) {
	group := loadFixture(t).ByCategory(Color)
	if !group.Has("red") || !group.Has("blue") || group.Len() != 2 {
		t.Errorf("keys = %v, want [red blue]", group.Keys())
	}
}

func TestByFamilyReturnsOnlyThatFamily(t *testing.T) {
	family := loadFixture(t).ByFamily(Color, "red").(*jsonv.Object)
	if !family.Has("500") || !family.Has("700") || family.Len() != 2 {
		t.Errorf("keys = %v, want [500 700]", family.Keys())
	}
}

func TestByFamilyWrapsATokenInItsName(t *testing.T) {
	store := NewStore()
	store.Load(parse(t, `{"color":{"black":{"$value":"#000","$type":"color"},"red":{"500":{"$value":"#f00","$type":"color"}}}}`))
	got := jsonv.Marshal(store.ByFamily(Color, "black"), "")
	if want := `{"black":{"$value":"#000","$type":"color"}}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestByFamilyUnknownIsEmptyObject(t *testing.T) {
	store := loadFixture(t)
	for _, name := range []string{"nope", "constructor", "toString", "__proto__"} {
		if got := jsonv.Marshal(store.ByFamily(Color, name), ""); got != "{}" {
			t.Errorf("ByFamily(%q) = %s, want {}", name, got)
		}
	}
}

func TestListWithoutFilterReturnsAllTokens(t *testing.T) {
	if n := len(loadFixture(t).List("")); n != 8 {
		t.Errorf("len = %d, want 8", n)
	}
}

func TestListByCategory(t *testing.T) {
	entries := loadFixture(t).List(Color)
	if len(entries) != 4 {
		t.Fatalf("len = %d, want 4", len(entries))
	}
	for _, e := range entries {
		if e.Category != Color {
			t.Errorf("category = %s, want color", e.Category)
		}
	}
}

func TestListFlattensPathsAndValues(t *testing.T) {
	entries := loadFixture(t).List(Spacing)
	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Path)
	}
	if want := []string{"spacing.scale.sm", "spacing.scale.md"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	if entries[0].Value != "8px" || entries[0].Type != "spacing" {
		t.Errorf("first entry = %+v", entries[0])
	}
}

func TestFamilies(t *testing.T) {
	got := loadFixture(t).Families(Color)
	slices.Sort(got)
	if want := []string{"blue", "red"}; !slices.Equal(got, want) {
		t.Errorf("families = %v, want %v", got, want)
	}
}

func TestIntegerKeysAreOrderedNumerically(t *testing.T) {
	store := NewStore()
	store.Load(parse(t, `{"spacing":{"s":{
		"b":{"$value":"1","$type":"x"},
		"10":{"$value":"2","$type":"x"},
		"2":{"$value":"3","$type":"x"}}}}`))
	var paths []string
	for _, e := range store.List(Spacing) {
		paths = append(paths, e.Path)
	}
	if want := []string{"spacing.s.2", "spacing.s.10", "spacing.s.b"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
}

func TestJSONUsesFixedCategoryOrderAndIgnoresOtherKeys(t *testing.T) {
	store := NewStore()
	store.Load(parse(t, `{"$schema":"x","spacing":{},"extra":{"a":1},"color":{}}`))
	want := "{\n  \"color\": {},\n  \"gradient\": {},\n  \"typography\": {},\n  \"spacing\": {}\n}"
	if got := store.JSON(""); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := store.JSON(Spacing), "{\n  \"spacing\": {}\n}"; got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReturnedTreesAreCopies(t *testing.T) {
	store := loadFixture(t)
	store.ByCategory(Color).Set("injected", "x")
	store.All().Set("injected", "x")
	if store.ByCategory(Color).Has("injected") {
		t.Error("mutating a returned tree changed the store")
	}
	if strings.Contains(store.JSON(""), "injected") {
		t.Error("mutating a returned tree changed the store")
	}
}

func TestLoadToleratesOddDocuments(t *testing.T) {
	store := NewStore()
	// Non-object roots and categories load as empty; strings inside groups
	// (such as a group-level "$description") are not tokens and are skipped.
	for _, doc := range []string{`[]`, `null`, `"x"`, `{"color":"x"}`, `{"color":[1]}`} {
		counts := store.Load(parse(t, doc))
		if n := counts[Color] + counts[Gradient] + counts[Typography] + counts[Spacing]; n != 0 {
			t.Errorf("Load(%s) counted %d tokens, want 0", doc, n)
		}
	}
	counts := store.Load(parse(t, `{"color":{"red":{"$description":"a group","50":{"$value":"#fff","$type":"color"}}}}`))
	if counts[Color] != 1 {
		t.Errorf("counted %d color tokens, want 1", counts[Color])
	}
	if n := len(store.List(Color)); n != 1 {
		t.Errorf("listed %d color tokens, want 1", n)
	}
}

func TestParseCategory(t *testing.T) {
	for _, c := range Categories {
		if got, ok := ParseCategory(string(c)); !ok || got != c {
			t.Errorf("ParseCategory(%q) = %q, %v", c, got, ok)
		}
	}
	for _, bad := range []string{"", "COLOR", "colors", "bogus"} {
		if _, ok := ParseCategory(bad); ok {
			t.Errorf("ParseCategory(%q) succeeded", bad)
		}
	}
}
