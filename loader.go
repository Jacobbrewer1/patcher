package patcher

import (
	"errors"
	"reflect"
)

var (
	// ErrInvalidType is returned when the provided type is not a pointer to a struct
	ErrInvalidType = errors.New("invalid type")

	// ErrInvalidFieldType is returned when the provided field type is not a struct
	ErrInvalidFieldType = errors.New("invalid field type")
)

// LoadDiff inserts the fields provided in the new object into the old object and returns the result.
//
// This can be if you are inserting a patch into an existing object but require a new object to be returned with
// all fields.
func LoadDiff[T any](old T, newT T) error {
	return loadDiff(old, newT)
}

func loadDiff[T any](old T, newT T) error {
	orv := reflect.ValueOf(old)
	if orv.Kind() != reflect.Ptr || orv.IsNil() {
		return ErrInvalidType
	}

	nrv := reflect.ValueOf(newT)
	if nrv.Kind() != reflect.Ptr || nrv.IsNil() {
		return ErrInvalidType
	}

	oElem := orv.Elem()
	nElem := nrv.Elem()

	if oElem.Kind() != reflect.Struct || nElem.Kind() != reflect.Struct {
		return ErrInvalidFieldType
	}

	for i := 0; i < oElem.NumField(); i++ {
		// Include only exported fields
		if !oElem.Field(i).CanSet() || !nElem.Field(i).CanSet() {
			continue
		}

		// Handle embedded structs (Anonymous fields)
		if oElem.Type().Field(i).Anonymous {
			// If the embedded field is a pointer, dereference it
			if oElem.Field(i).Kind() == reflect.Ptr {
				if !oElem.Field(i).IsNil() && !nElem.Field(i).IsNil() {
					if err := loadDiff(oElem.Field(i).Interface(), nElem.Field(i).Interface()); err != nil {
						return err
					}
				} else if nElem.Field(i).IsValid() && !nElem.Field(i).IsNil() {
					oElem.Field(i).Set(nElem.Field(i))
				}

				continue
			}

			if err := loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If the field is a struct, we need to recursively call LoadDiff
		if oElem.Field(i).Kind() == reflect.Struct {
			if err := loadDiff(oElem.Field(i).Addr().Interface(), nElem.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided. We ignore zero values as they are not
		// provided in the new object.
		if !nElem.Field(i).IsZero() {
			oElem.Field(i).Set(nElem.Field(i))
		}
	}

	return nil
}
