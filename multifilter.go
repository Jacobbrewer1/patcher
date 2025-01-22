package patcher

import "strings"

type MultiFilter interface {
	Wherer
	Add(where Wherer)
}

type multiFilter struct {
	whereSql  *strings.Builder
	whereArgs []any
}

func NewMultiFilter() MultiFilter {
	return &multiFilter{
		whereSql:  new(strings.Builder),
		whereArgs: nil,
	}
}

func (m *multiFilter) Add(where Wherer) {
	appendWhere(where, m.whereSql, &m.whereArgs)
}

func (m *multiFilter) Where() (string, []any) {
	return m.whereSql.String(), m.whereArgs
}
