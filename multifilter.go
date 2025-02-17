package patcher

import "strings"

type Filter interface {
	Joiner
	Wherer
}

type MultiFilter interface {
	Filter
	Add(filter any)
}

type multiFilter struct {
	joinSql   *strings.Builder
	joinArgs  []any
	whereSql  *strings.Builder
	whereArgs []any
}

func (m *multiFilter) Join() (sqlStr string, args []any) {
	return m.joinSql.String(), m.joinArgs
}

func (m *multiFilter) Where() (sqlStr string, args []any) {
	return m.whereSql.String(), m.whereArgs
}

func (m *multiFilter) Add(filter any) {
	if joiner, ok := filter.(Joiner); ok {
		appendJoin(joiner, m.joinSql, &m.joinArgs)
	}

	if wherer, ok := filter.(Wherer); ok {
		appendWhere(wherer, m.whereSql, &m.whereArgs)
	}
}

func NewMultiFilter() MultiFilter {
	return &multiFilter{
		joinSql:   new(strings.Builder),
		joinArgs:  nil,
		whereSql:  new(strings.Builder),
		whereArgs: nil,
	}
}
