package patcher

import (
	"errors"
	"reflect"
)

// ptr returns a pointer to the value passed in.
func ptr[T any](v T) *T {
	return &v
}

func isPointerToStruct[T any](t T) bool {
	rv := reflect.ValueOf(t)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false
	}

	return rv.Elem().Kind() == reflect.Struct
}

// IgnoreNoChangesErr ignores the ErrNoChanges error. This is useful when you want to ignore the error when no changes
// were made. Please ensure that you are still handling the errors as needed. We will return a "nil" patch when there
// are no changes as the ErrNoChanges error is returned.
func IgnoreNoChangesErr(err error) error {
	switch {
	case errors.Is(err, ErrNoChanges):
		return nil
	default:
		return err
	}
}
