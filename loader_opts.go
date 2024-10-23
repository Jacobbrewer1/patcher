package patcher

import "strings"

type LoaderOption func(*loader)

// WithIncludeZeroValues sets whether zero values should be included in the patch.
//
// This is useful when you want to set a field to zero.
func WithIncludeZeroValues() func(*loader) {
	return func(l *loader) {
		l.includeZeroValues = true
	}
}

// WithIncludeNilValues sets whether nil values should be included in the patch.
//
// This is useful when you want to set a field to nil.
func WithIncludeNilValues() func(*loader) {
	return func(l *loader) {
		l.includeNilValues = true
	}
}

// WithIgnoredFields sets the fields to ignore when patching.
//
// This should be the actual field name, not the JSON tag name or the db tag name.
//
// Note. When we parse the slice of strings, we convert them to lowercase to ensure that the comparison is
// case-insensitive.
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
//
// Note. The field name is wrapped with `strings.ToLower` before being passed to this function, so please ensure that
// the field name is in lowercase if you are comparing it with this function.
func WithIgnoredFieldsFunc(f func(fieldName string, oldValue, newValue any) bool) func(*loader) {
	return func(l *loader) {
		if f == nil {
			return
		}

		l.ignoreFieldsFunc = f
	}
}
