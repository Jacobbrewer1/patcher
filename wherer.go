package patcher

type Wherer interface {
	Where() (string, []any)
}
