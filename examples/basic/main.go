package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID   *int    `db:"-" json:"id,omitempty"`
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

func main() {
	const jsonStr = `{"id": 1, "name": "john"}`

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

	// Output:
	// UPDATE people
	// SET name = ?
	// WHERE (1 = 1)
	//   AND (
	//     id = ?
	//     )
	fmt.Println(sqlStr)

	// Output:
	// ["John", 1]
	fmt.Println(args)
}
