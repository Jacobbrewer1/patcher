package patcher

// Joiner is an interface that can be used to specify the JOIN clause to use when the SQL is being generated.
type Joiner interface {
	Join() (string, []any)
}
