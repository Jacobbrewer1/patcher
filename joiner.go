package patcher

import "strings"

// Joiner is an interface that can be used to specify the JOIN clause to use when the SQL is being generated.
type Joiner interface {
	Join() (string, []any)
}

func appendJoin(join Joiner, builder *strings.Builder, args *[]any) {
	if join == nil {
		return
	}
	jSQL, jArgs := join.Join()
	if jArgs == nil {
		jArgs = make([]any, 0)
	}
	builder.WriteString(strings.TrimSpace(jSQL))
	builder.WriteString("\n")
	*args = append(*args, jArgs...)
}

type joinStringOption struct {
	join string
	args []any
}

func (j *joinStringOption) Join() (sqlStr string, args []any) {
	return j.join, j.args
}
