package patcher

import (
	"errors"
	"reflect"
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
func LoadDiff[T any](old, newT *T, opts ...PatchOpt) error {
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

	for i := range oElem.NumField() {
		oField := oElem.Field(i)
		nField := nElem.Field(i)

		// Include only exported fields
		if !oField.CanSet() || !nField.CanSet() {
			continue
		}

		oldField := oElem.Type().Field(i)

		// Handle embedded structs (Anonymous fields)
		if oldField.Anonymous {
			if err := s.handleEmbeddedStruct(oField, nField); err != nil {
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
		if s.checkSkipField(&oldField) {
			continue
		}

		patcherOptsTag := oldField.Tag.Get(TagOptsName)

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided based on the configuration.
		if nField.Kind() != reflect.Ptr && (!nField.IsZero() || s.shouldIncludeZero(patcherOptsTag)) {
			oField.Set(nField)
		} else if nField.Kind() == reflect.Ptr && (!nField.IsNil() || s.shouldIncludeNil(patcherOptsTag)) {
			oField.Set(nField)
		}
	}

	return nil
}

func (s *SQLPatch) handleEmbeddedStruct(oField, nField reflect.Value) error {
	if oField.Kind() != reflect.Ptr {
		return s.loadDiff(oField.Addr().Interface(), nField.Addr().Interface())
	}

	if !oField.IsNil() && !nField.IsNil() {
		return s.loadDiff(oField.Interface(), nField.Interface())
	} else if nField.IsValid() && !nField.IsNil() {
		oField.Set(nField)
	}

	return nil
}
