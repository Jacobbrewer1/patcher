package patcher

import "reflect"

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
