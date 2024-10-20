package main

import (
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Something struct {
	Number       int
	Text         string
	PrePopulated string
	NewText      string
}

type SomeWhere struct {
	id int
}

func NewSomeWhere(id int) *SomeWhere {
	return &SomeWhere{id: id}
}

func (s *SomeWhere) Where() (string, []any) {
	return "id = ?", []any{s.id}
}

func main() {
	s := Something{
		Number:       5,
		Text:         "Old Text",
		PrePopulated: "PrePopulated",
		NewText:      "New Text",
	}

	n := Something{
		Number:       5,
		Text:         "Old Text",
		PrePopulated: "PrePopulatedDifferent",
		NewText:      "New Text",
	}

	wherer := NewSomeWhere(5)

	// The patcher.LoadDiff function will apply the changes from n to s.
	patch, err := patcher.NewDiffSQLPatch(
		&s,
		&n,
		patcher.WithTable("table_name"),
		patcher.WithWhere(wherer),
	)
	if err != nil {
		panic(err)
	}

	sqlStr, sqlArgs, err := patch.GenerateSQL()
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(sqlArgs)
}
