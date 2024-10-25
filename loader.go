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

type loader struct {
	// includeZeroValues determines whether zero values should be included in the patch
	includeZeroValues bool

	// includeNilValues determines whether nil values should be included in the patch
	includeNilValues bool

	// ignoreFields is a list of fields to ignore when patching
	ignoreFields []string

	// ignoreFieldsFunc is a function that determines whether a field should be ignored
	//
	// This func should return true is the field is to be ignored
	ignoreFieldsFunc func(fieldName string, oldValue, newValue any) bool
}

func newLoader(opts ...LoaderOption) *loader {
	// Default options
	l := &loader{
		includeZeroValues: false,
		includeNilValues:  false,
		ignoreFields:      nil,
		ignoreFieldsFunc:  nil,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
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
func LoadDiff[T any](old *T, newT *T, opts ...LoaderOption) error {
	return newLoader(opts...).loadDiff(old, newT)
}

func (l *loader) loadDiff(old, newT any) error {
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
					if err := l.loadDiff(oElem.Field(i).Interface(), nElem.Field(i).Interface()); err != nil {
						return err
					}
				} else if nElem.Field(i).IsValid() && !nElem.Field(i).IsNil() {
					oElem.Field(i).Set(nElem.Field(i))
				}

				continue
			}

			if err := l.loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If the field is a struct, we need to recursively call LoadDiff
		if oElem.Field(i).Kind() == reflect.Struct {
			if err := l.loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// See if the field should be ignored.
		if l.checkSkipField(oElem.Type().Field(i), oElem.Field(i).Interface(), nElem.Field(i).Interface()) {
			continue
		}

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided based on the configuration.
		if !nElem.Field(i).IsZero() || l.includeZeroValues {
			oElem.Field(i).Set(nElem.Field(i))
		} else if nElem.Field(i).Kind() == reflect.Ptr && nElem.Field(i).IsNil() && l.includeNilValues {
			oElem.Field(i).Set(nElem.Field(i))
		}
	}

	return nil
}

func (l *loader) checkSkipField(field reflect.StructField, oldValue, newValue any) bool {
	// The ignore fields tag takes precedence over the ignore fields list
	if l.checkSkipTag(field) {
		return true
	}

	return l.ignoredFieldsCheck(strings.ToLower(field.Name), oldValue, newValue)
}

func (l *loader) checkSkipTag(field reflect.StructField) bool {
	val, ok := field.Tag.Lookup(TagOptsName)
	if !ok {
		return false
	}

	tags := strings.Split(val, TagOptSeparator)
	return slices.Contains(tags, TagOptSkip)
}

func (l *loader) ignoredFieldsCheck(field string, oldValue, newValue any) bool {
	return l.checkIgnoredFields(field) || l.checkIgnoreFunc(field, oldValue, newValue)
}

func (l *loader) checkIgnoreFunc(field string, oldValue, newValue any) bool {
	return l.ignoreFieldsFunc != nil && l.ignoreFieldsFunc(strings.ToLower(field), oldValue, newValue)
}

func (l *loader) checkIgnoredFields(field string) bool {
	return len(l.ignoreFields) > 0 && slices.Contains(l.ignoreFields, strings.ToLower(field))
}
