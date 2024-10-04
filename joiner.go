package patcher

type Joiner interface {
	Join() (string, []any)
}
