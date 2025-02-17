package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID    *int    `db:"id"`
	Name  *string `db:"name"`
	Age   *int    `db:"age"`
	Email *string `db:"email"`
}

type PersonFilter struct {
	ID    *int    `db:"id"`
	Email *string `db:"email"`
}

func NewPersonFilter(id int, email string) patcher.Filter {
	return &PersonFilter{
		ID:    &id,
		Email: &email,
	}
}

func (p *PersonFilter) Where() (string, []interface{}) {
	return "id = ?", []interface{}{*p.ID}
}

func (p *PersonFilter) Join() (string, []any) {
	return "JOIN contacts c ON c.person_id = p.id AND c.email = ?", []any{*p.Email}
}

func main() {
	const jsonStr = `{"name": "john", "age": 25, "email": "john@exampletwo.com"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonFilter(1, *person.Email)

	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithFilter(condition),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(args)
}
