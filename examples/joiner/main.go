package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID    *int    `db:"id" json:"id,omitempty"`
	Name  *string `db:"name" json:"name,omitempty"`
	Age   *int    `db:"age" json:"age,omitempty"`
	Email *string `db:"email" json:"email,omitempty"`
}

type PersonWhere struct {
	ID *int `db:"id"`
}

func NewPersonWhere(id int) patcher.Wherer {
	return &PersonWhere{
		ID: &id,
	}
}

func (p *PersonWhere) Where() (sqlStr string, sqlArgs []any) {
	return "id = ?", []any{*p.ID}
}

type PersonContactJoiner struct {
	Email *string
}

func NewPersonContactJoiner(email string) patcher.Joiner {
	return &PersonContactJoiner{
		Email: &email,
	}
}

func (p *PersonContactJoiner) Join() (sqlStr string, sqlArgs []any) {
	return "JOIN contacts c ON c.person_id = p.id AND c.email = ?", []any{*p.Email}
}

func main() {
	const jsonStr = `{"name": "john", "age": 25, "email": "john@exampletwo.com"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(1)
	joiner := NewPersonContactJoiner(*person.Email)

	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithWhere(condition),
		patcher.WithJoin(joiner),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(args)
}
