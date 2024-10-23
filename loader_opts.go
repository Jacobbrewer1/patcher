package patcher

import "strings"

type loaderOption func(*loader)

func WithIncludeZeroValues() func(*loader) {
	return func(l *loader) {
		l.includeZeroValues = true
	}
}

func WithIncludeNilValues() func(*loader) {
	return func(l *loader) {
		l.includeNilValues = true
	}
}

func WithIncludeEmptyValues() func(*loader) {
	return func(l *loader) {
		l.includeEmptyValues = true
	}
}

// WithIgnoredFields sets the fields to ignore when patching.
//
// This should be the actual field name, not the JSON tag name or the db tag name.
func WithIgnoredFields(fields ...string) func(*loader) {
	return func(l *loader) {
		if len(fields) == 0 {
			return
		}

		for i := range fields {
			fields[i] = strings.ToLower(fields[i])
		}

		l.ignoreFields = fields
	}
}

// WithIgnoredFieldsFunc sets a function that determines whether a field should be ignored when patching.
func WithIgnoredFieldsFunc(f func(string) bool) func(*loader) {
	return func(l *loader) {
		if f == nil {
			return
		}

		l.ignoreFieldsFunc = f
	}
}
