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

// LoadDiff inserts the fields from the new struct pointer into the old struct pointer, updating the old struct.
//
// Note that it only updates non-zero value fields, meaning you cannot set any field to zero, the empty string, etc.
// This behavior is configurable by setting the includeZeroValues option to true or for nil values by setting includeNilValues.
// Please see the LoaderOption's for more configuration options.
//
// This function is useful if you are inserting a patch into an existing object but require a new object to be returned with
// all fields updated.
func LoadDiff[T any](old *T, newT *T, opts ...PatchOpt) error {
	return newPatchDefaults(opts...).loadDiff(old, newT)
}

// loadDiff inserts the fields provided in the new struct pointer into the old struct pointer and injects the new
// values into the old struct. It only pushes non-zero value updates, meaning you cannot set any field to zero,
// the empty string, etc. This is configurable by setting the includeZeroValues option to true or for nil values
// by setting includeNilValues. It handles embedded structs and recursively calls loadDiff for nested structs.
func (s *SQLPatch) loadDiff(old, newT any) error {
	if !isPointerToStruct(old) || !isPointerToStruct(newT) {
		return ErrInvalidType
	}

	oElem := reflect.ValueOf(old).Elem()
	nElem := reflect.ValueOf(newT).Elem()

	for i := 0; i < oElem.NumField(); i++ {
		oField := oElem.Field(i)
		nField := nElem.Field(i)

		// Include only exported fields
		if !oField.CanSet() || !nField.CanSet() {
			continue
		}

		// Handle embedded structs (Anonymous fields)
		if oElem.Type().Field(i).Anonymous {
			// If the embedded field is a pointer, dereference it
			if oField.Kind() == reflect.Ptr {
				if !oField.IsNil() && !nField.IsNil() { // If both are not nil, we need to recursively call LoadDiff
					if err := s.loadDiff(oField.Interface(), nField.Interface()); err != nil {
						return err
					}
				} else if nElem.Field(i).IsValid() && !nField.IsNil() {
					oField.Set(nField)
				}

				continue
			}

			if err := s.loadDiff(oField.Addr().Interface(), nField.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If the field is a struct, we need to recursively call LoadDiff
		if oField.Kind() == reflect.Struct {
			if err := s.loadDiff(oField.Addr().Interface(), nField.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// See if the field should be ignored.
		if s.checkSkipField(oElem.Type().Field(i)) {
			continue
		}

		patcherOptsTag := oElem.Type().Field(i).Tag.Get(TagOptsName)

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided based on the configuration.
		if nElem.Field(i).Kind() != reflect.Ptr && (!nField.IsZero() || s.shouldIncludeZero(patcherOptsTag)) {
			oElem.Field(i).Set(nElem.Field(i))
		} else if nElem.Field(i).Kind() == reflect.Ptr && (!nField.IsNil() || s.shouldIncludeNil(patcherOptsTag)) {
			oField.Set(nElem.Field(i))
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
