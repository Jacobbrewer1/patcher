package main

import (
	"encoding/json"
	"fmt"

	"github.com/Jacobbrewer1/patcher"
)

type Person struct {
	ID    *int    `db:"id"`
	Name  *string `db:"name"`
	Age   *int    `db:"age"`
	Email *string `db:"email"`
}

type PersonWhere struct {
	ID *int `db:"id"`
}

func NewPersonWhere(id int) *PersonWhere {
	return &PersonWhere{
		ID: &id,
	}
}

func (p *PersonWhere) Where() (string, []interface{}) {
	return "id = ?", []interface{}{*p.ID}
}

type ContactWhere struct {
	Email *string
}

func NewContactWhere(email string) *ContactWhere {
	return &ContactWhere{
		Email: &email,
	}
}

func (c *ContactWhere) Where() (string, []interface{}) {
	return "email = ?", []interface{}{*c.Email}
}

type PersonContactJoiner struct {
	Email *string
}

func NewPersonContactJoiner(email string) *PersonContactJoiner {
	return &PersonContactJoiner{
		Email: &email,
	}
}

func (p *PersonContactJoiner) Join() (string, []any) {
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
