package patcher

import (
	"errors"
	"reflect"
)

// LoadDiff inserts the fields provided in the new object into the old object and returns the result.
//
// This can be if you are inserting a patch into an existing object but require a new object to be returned with
// all fields.
func LoadDiff[T any](old T, new T) error {
	return loadDiff(old, new)
}

func loadDiff[T any](old T, new T) error {
	orv := reflect.ValueOf(old)
	if orv.Kind() != reflect.Ptr || orv.IsNil() {
		return errors.New("old must be a pointer")
	}

	nrv := reflect.ValueOf(new)
	if nrv.Kind() != reflect.Ptr || nrv.IsNil() {
		return errors.New("new must be a pointer")
	}

	oElem := orv.Elem()
	nElem := nrv.Elem()

	for i := 0; i < orv.Elem().NumField(); i++ {
		if !oElem.Field(i).CanSet() || !nElem.Field(i).CanSet() {
			continue
		}

		// Compare the old and new fields.
		//
		// New fields take priority over old fields if they are provided. We ignore zero values as they are not
		// provided in the new object.
		if nElem.Field(i).IsZero() {
			nElem.Field(i).Set(oElem.Field(i))
		}

		// Remove the value from the old object. Nil the value if it is a pointer.
		if oElem.Field(i).Kind() == reflect.Ptr {
			oElem.Field(i).Set(reflect.Zero(oElem.Field(i).Type()))
		} else {
			oElem.Field(i).Set(reflect.Zero(oElem.Field(i).Type()))
		}
	}

	return nil
}
