package patcher

// ptr returns a pointer to the value passed in.
func ptr[T any](v T) *T {
	return &v
}
