package patcher

import "strings"

// Wherer is an interface that can be used to specify the WHERE clause to use. By using this interface,
// the package will default to using an "AND" WHERE clause. If you want to use an "OR" WHERE clause, you can
// use the WhereTyper interface instead.
type Wherer interface {
	Where() (string, []any)
}

// WhereTyper is an interface that can be used to specify the type of WHERE clause to use. By using this
// interface, you can specify whether to use an "AND" or "OR" WHERE clause.
type WhereTyper interface {
	Wherer
	WhereType() WhereType
}

type WhereType string

const (
	WhereTypeAnd WhereType = "AND"
	WhereTypeOr  WhereType = "OR"
)

func (w WhereType) IsValid() bool {
	switch w {
	case WhereTypeAnd, WhereTypeOr:
		return true
	}
	return false
}

func appendWhere(where Wherer, builder *strings.Builder, args *[]any) {
	if where == nil {
		return
	}
	wSQL, fwArgs := where.Where()
	if fwArgs == nil {
		fwArgs = make([]any, 0)
	}
	wtStr := WhereTypeAnd // default to AND
	wt, ok := where.(WhereTyper)
	if ok && wt.WhereType().IsValid() {
		wtStr = wt.WhereType()
	}
	builder.WriteString(string(wtStr) + " ")
	builder.WriteString(strings.TrimSpace(wSQL))
	builder.WriteString("\n")
	*args = append(*args, fwArgs...)
}
