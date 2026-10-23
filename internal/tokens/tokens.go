// Package tokens loads and queries DTCG design tokens.
package tokens

import (
	"strconv"

	"github.com/pglevy/mint/internal/jsonv"
)

// Category is a top-level token category.
type Category string

const (
	Color      Category = "color"
	Gradient   Category = "gradient"
	Typography Category = "typography"
	Spacing    Category = "spacing"
)

// Categories lists the valid categories in display order.
var Categories = []Category{Color, Gradient, Typography, Spacing}

// ParseCategory reports whether s names a valid category.
func ParseCategory(s string) (Category, bool) {
	for _, c := range Categories {
		if string(c) == s {
			return c, true
		}
	}
	return "", false
}

// CategoryNames returns the valid category names as strings.
func CategoryNames() []string {
	names := make([]string, len(Categories))
	for i, c := range Categories {
		names[i] = string(c)
	}
	return names
}

// Entry is a single flattened token with its resolved path and value.
type Entry struct {
	Path     string
	Value    jsonv.Value
	Type     jsonv.Value // a string in practice; kept as-is so output round-trips
	Category Category
}

// JSON returns the entry as the object printed by `list --json`.
func (e Entry) JSON() *jsonv.Object {
	o := jsonv.NewObject()
	o.Set("path", e.Path)
	o.Set("value", e.Value)
	o.Set("type", e.Type)
	o.Set("category", string(e.Category))
	return o
}

// Counts holds the number of tokens loaded per category.
type Counts map[Category]int

// isToken reports whether node is a DTCG token (an object with $value and $type).
func isToken(node jsonv.Value) (*jsonv.Object, bool) {
	o, ok := node.(*jsonv.Object)
	if !ok || !o.Has("$value") || !o.Has("$type") {
		return nil, false
	}
	return o, true
}

// children yields the named children of a group. Objects yield their
// properties; arrays yield their elements keyed by index, as JavaScript's
// Object.entries does. Anything else (strings, numbers) is a leaf and has no
// children.
func children(node jsonv.Value, yield func(key string, child jsonv.Value)) {
	switch n := node.(type) {
	case *jsonv.Object:
		for _, k := range n.Keys() {
			child, _ := n.Get(k)
			yield(k, child)
		}
	case []jsonv.Value:
		for i, child := range n {
			yield(strconv.Itoa(i), child)
		}
	}
}

func collect(node jsonv.Value, prefix string, category Category, out *[]Entry) {
	children(node, func(key string, child jsonv.Value) {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if tok, ok := isToken(child); ok {
			value, _ := tok.Get("$value")
			typ, _ := tok.Get("$type")
			*out = append(*out, Entry{Path: path, Value: value, Type: typ, Category: category})
			return
		}
		collect(child, path, category, out)
	})
}

func count(node jsonv.Value) int {
	n := 0
	children(node, func(_ string, child jsonv.Value) {
		if _, ok := isToken(child); ok {
			n++
		} else {
			n += count(child)
		}
	})
	return n
}

// Store holds the tokens of each category, in their original order.
type Store struct {
	tree map[Category]*jsonv.Object
}

// NewStore returns an empty Store.
func NewStore() *Store {
	s := &Store{}
	s.reset()
	return s
}

func (s *Store) reset() {
	s.tree = map[Category]*jsonv.Object{}
	for _, c := range Categories {
		s.tree[c] = jsonv.NewObject()
	}
}

// Load replaces the store's contents with the categories found in root, a
// parsed tokens document, and returns the token count per category. Keys
// other than the four categories (such as "$schema") are ignored.
func (s *Store) Load(root jsonv.Value) Counts {
	s.reset()
	counts := Counts{}
	if doc, ok := root.(*jsonv.Object); ok {
		for _, c := range Categories {
			if group, ok := doc.Get(string(c)); ok {
				if g, ok := group.(*jsonv.Object); ok {
					s.tree[c] = g
				}
			}
		}
	}
	for _, c := range Categories {
		counts[c] = count(s.tree[c])
	}
	return counts
}

// All returns a copy of every category's tokens, keyed by category.
func (s *Store) All() *jsonv.Object {
	return s.document(Categories)
}

// ByCategory returns a copy of one category's token tree.
func (s *Store) ByCategory(c Category) *jsonv.Object {
	return jsonv.Clone(s.tree[c]).(*jsonv.Object)
}

// ByFamily returns a copy of one top-level family of a category. If the
// family is itself a token it is returned wrapped in an object keyed by its
// name; an unknown family yields an empty object.
func (s *Store) ByFamily(c Category, family string) jsonv.Value {
	child, ok := s.tree[c].Get(family)
	if !ok {
		return jsonv.NewObject()
	}
	if _, isTok := isToken(child); isTok {
		wrapped := jsonv.NewObject()
		wrapped.Set(family, jsonv.Clone(child))
		return wrapped
	}
	return jsonv.Clone(child)
}

// Families returns the top-level family names of a category.
func (s *Store) Families(c Category) []string {
	return append([]string(nil), s.tree[c].Keys()...)
}

// List flattens the tokens of one category, or of all categories when c is
// empty.
func (s *Store) List(c Category) []Entry {
	cats := Categories
	if c != "" {
		cats = []Category{c}
	}
	var entries []Entry
	for _, cat := range cats {
		collect(s.tree[cat], string(cat), cat, &entries)
	}
	return entries
}

// JSON renders the tokens as indented DTCG JSON, for one category or, when c
// is empty, all of them.
func (s *Store) JSON(c Category) string {
	if c != "" {
		return jsonv.Marshal(s.document([]Category{c}), "  ")
	}
	return jsonv.Marshal(s.document(Categories), "  ")
}

func (s *Store) document(cats []Category) *jsonv.Object {
	doc := jsonv.NewObject()
	for _, c := range cats {
		doc.Set(string(c), jsonv.Clone(s.tree[c]))
	}
	return doc
}
