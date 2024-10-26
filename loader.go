package patcher

import (
	"errors"
	"reflect"
	"slices"
	"strings"
)

var (
	// ErrInvalidType is returned when the provided type is not a pointer to a struct
	ErrInvalidType = errors.New("invalid type: must pointer to struct")
)

func NewPatch(opts ...PatchOpt) *SQLPatch {
	// Default options
	p := &SQLPatch{
		fields:            nil,
		args:              nil,
		db:                nil,
		tagName:           "",
		table:             "",
		where:             new(strings.Builder),
		whereArgs:         nil,
		joinSql:           new(strings.Builder),
		joinArgs:          nil,
		includeZeroValues: false,
		includeNilValues:  false,
		ignoreFields:      nil,
		ignoreFieldsFunc:  nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// LoadDiff inserts the fields provided in the new struct pointer into the old struct pointer and injects the new
// values into the old struct
//
// Note that it only pushes non-zero value updates, meaning you cannot set any field to zero, the empty string, etc.
// This is configurable by setting the includeZeroValues option to true or for nil values by setting includeNilValues.
// Please see the LoaderOption's for more configuration options.
//
// This can be if you are inserting a patch into an existing object but require a new object to be returned with
// all fields
func LoadDiff[T any](old *T, newT *T, opts ...PatchOpt) error {
	return NewPatch(opts...).loadDiff(old, newT)
}

func (s *SQLPatch) loadDiff(old, newT any) error {
	if !isPointerToStruct(old) || !isPointerToStruct(newT) {
		return ErrInvalidType
	}

	oElem := reflect.ValueOf(old).Elem()
	nElem := reflect.ValueOf(newT).Elem()

	for i := 0; i < oElem.NumField(); i++ {
		// Include only exported fields
		if !oElem.Field(i).CanSet() || !nElem.Field(i).CanSet() {
			continue
		}

		// Handle embedded structs (Anonymous fields)
		if oElem.Type().Field(i).Anonymous {
			// If the embedded field is a pointer, dereference it
			if oElem.Field(i).Kind() == reflect.Ptr {
				if !oElem.Field(i).IsNil() && !nElem.Field(i).IsNil() { // If both are not nil, we need to recursively call LoadDiff
					if err := s.loadDiff(oElem.Field(i).Interface(), nElem.Field(i).Interface()); err != nil {
						return err
					}
				} else if nElem.Field(i).IsValid() && !nElem.Field(i).IsNil() {
					oElem.Field(i).Set(nElem.Field(i))
				}

				continue
			}

			if err := s.loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If the field is a struct, we need to recursively call LoadDiff
		if oElem.Field(i).Kind() == reflect.Struct {
			if err := s.loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// See if the field should be ignored.
		if s.checkSkipField(oElem.Type().Field(i)) {
			continue
		}

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided based on the configuration.
		if nElem.Field(i).Kind() != reflect.Ptr && (!nElem.Field(i).IsZero() || s.includeZeroValues) {
			oElem.Field(i).Set(nElem.Field(i))
		} else if nElem.Field(i).Kind() == reflect.Ptr && (!nElem.Field(i).IsNil() || s.includeNilValues) {
			oElem.Field(i).Set(nElem.Field(i))
		}
	}

	return nil
}

func (s *SQLPatch) checkSkipField(field reflect.StructField) bool {
	// The ignore fields tag takes precedence over the ignore fields list
	if s.checkSkipTag(field) {
		return true
	}

	return s.ignoredFieldsCheck(field)
}

func (s *SQLPatch) checkSkipTag(field reflect.StructField) bool {
	val, ok := field.Tag.Lookup(TagOptsName)
	if !ok {
		return false
	}

	tags := strings.Split(val, TagOptSeparator)
	return slices.Contains(tags, TagOptSkip)
}

func (s *SQLPatch) ignoredFieldsCheck(field reflect.StructField) bool {
	return s.checkIgnoredFields(strings.ToLower(field.Name)) || s.checkIgnoreFunc(field)
}

func (s *SQLPatch) checkIgnoreFunc(field reflect.StructField) bool {
	return s.ignoreFieldsFunc != nil && s.ignoreFieldsFunc(field)
}

func (s *SQLPatch) checkIgnoredFields(field string) bool {
	return len(s.ignoreFields) > 0 && slices.Contains(s.ignoreFields, strings.ToLower(field))
}
