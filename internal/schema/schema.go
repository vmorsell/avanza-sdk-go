// Package schema detects drift between live JSON API responses and the Go
// structs that model them. It reports JSON keys present in a response that have
// no matching struct field, which is how new or renamed upstream fields are
// discovered without hand-diffing payloads.
package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var jsonUnmarshalerType = reflect.TypeFor[json.Unmarshaler]()

// UnknownFields returns the dotted paths of every JSON key in raw that has no
// matching field in target's type. target must be a pointer to the struct the
// response is expected to decode into.
//
// Paths use "." for nested objects and "[]" for array elements (deduplicated
// across elements, so a field missing from every element is reported once).
// Fields whose type decodes itself via a custom json.Unmarshaler (e.g.
// json.RawMessage) are opaque: their contents are never walked.
//
// UnknownFields only reports keys the struct lacks; it does not report struct
// fields absent from the response. Type mismatches (e.g. a value that changed
// from number to object) are not reported here — a strict json.Unmarshal into
// the same target surfaces those.
func UnknownFields(raw []byte, target any) ([]string, error) {
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("unmarshal raw: %w", err)
	}
	t := reflect.TypeOf(target)
	if t == nil {
		return nil, fmt.Errorf("target must be a non-nil pointer to a struct")
	}

	set := map[string]struct{}{}
	walk(data, t, "", set)

	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

func walk(data any, t reflect.Type, path string, set map[string]struct{}) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch d := data.(type) {
	case map[string]any:
		switch t.Kind() {
		case reflect.Struct:
			fields := structFields(t)
			for key, val := range d {
				f, ok := fields[strings.ToLower(key)]
				if !ok {
					set[join(path, key)] = struct{}{}
					continue
				}
				if isOpaque(f.Type) {
					continue // field decodes itself; its shape isn't modelled by struct fields
				}
				walk(val, f.Type, join(path, key), set)
			}
		case reflect.Map:
			// Keys are dynamic and therefore all "known"; recurse into values.
			for key, val := range d {
				walk(val, t.Elem(), join(path, key), set)
			}
		default:
			// JSON object modelled by a non-struct Go type (e.g. any). Treat as
			// opaque; a strict Unmarshal catches genuine type mismatches.
		}
	case []any:
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			for _, elem := range d {
				walk(elem, t.Elem(), path+"[]", set)
			}
		}
	default:
		// Scalar leaf (string, number, bool, null) — nothing to check.
	}
}

// structFields maps the lowercased JSON name of every exported, serialized
// field to its StructField. Lowercasing mirrors encoding/json's case-insensitive
// key matching so casing differences aren't mistaken for drift.
func structFields(t reflect.Type) map[string]reflect.StructField {
	m := make(map[string]reflect.StructField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue // unexported
		}
		name := jsonName(f)
		if name == "-" {
			continue // explicitly not serialized
		}
		m[strings.ToLower(name)] = f
	}
	return m
}

// isOpaque reports whether a field decodes itself via a custom json.Unmarshaler
// (json.RawMessage, time.Time, and the like). Such a field's contents are not
// modelled by reflectable struct fields, so the walker must not descend into it.
// The pointer check mirrors encoding/json, which uses the addressable (pointer)
// value — json.RawMessage's UnmarshalJSON has a pointer receiver.
func isOpaque(t reflect.Type) bool {
	return t.Implements(jsonUnmarshalerType) || reflect.PointerTo(t).Implements(jsonUnmarshalerType)
}

func jsonName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	if name := strings.Split(tag, ",")[0]; name != "" {
		return name
	}
	return f.Name
}

func join(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}
