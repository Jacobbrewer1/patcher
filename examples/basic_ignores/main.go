package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID           *int    `json:"id" db:"id"`
	Name         *string `json:"name" db:"name"`
	IgnoredField *string `json:"ignored_field" patcher:"-"`
}

type PersonWhere struct {
	ID *int `db:"id"`
}

func NewPersonWhere(id int) *PersonWhere {
	return &PersonWhere{
		ID: &id,
	}
}

func (p *PersonWhere) Where() (sqlStr string, sqlArgs []any) {
	return "id = ?", []any{*p.ID}
}

func main() {
	const jsonStr = `{"id": 1, "name": "john", "ignored_field": "ignored"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(*person.ID)

	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithWhere(condition),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(args)
}
