package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID   *int    `db:"id" json:"id,omitempty"`
	Name *string `db:"name" json:"name,omitempty"`
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

func (p *PersonWhere) WhereType() patcher.WhereType {
	// Please do not use this in production code. This is just an example.
	if *p.ID == 1 {
		return patcher.WhereTypeOr
	}

	return patcher.WhereTypeAnd
}

func main() {
	const jsonStr = `{"name": "john"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(0)
	conditionOr := NewPersonWhere(1)

	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithWhere(condition),
		patcher.WithWhere(conditionOr),
	)
	if err != nil {
		panic(err)
	}

	// The where clause here should be "id = ?" and the where type should be "OR"
	fmt.Println(sqlStr)
	fmt.Println(args)
}
