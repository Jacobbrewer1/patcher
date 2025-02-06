package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID                *int    `db:"id"`
	Name              *string `db:"name"`
	IgnoredByTag      string  `db:"-"`
	IncludedZeroValue string  `db:"includedZeroValue"`
	IncludedNilValue  string  `db:"includedNilValue"`
	IgnoredByFunc     string  `db:"ignoredFieldByFunc"`
	IgnoredByList     string  `db:"ignoredFieldByList"`
}

type PersonWhere struct {
	ID *int `db:"id"`
}

func NewPersonWhere(id int) *PersonWhere {
	return &PersonWhere{
		ID: &id,
	}
}

func (p *PersonWhere) Where() (string, []any) {
	return "id = ?", []any{*p.ID}
}

func main() {
	const jsonStr = `{"id": 1, "name": "john", "ignoredField": "ignored", "ignoredNilField": null, "ignoredFieldByFunc": "ignored", "ignoredFieldByList": "ignored"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(*person.ID)

	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithWhere(condition),
		patcher.WithIncludeZeroValues(),
		patcher.WithIncludeNilValues(),
		patcher.WithIgnoredFields("ignoredbylist"),
		patcher.WithIgnoredFieldsFunc(func(field *reflect.StructField) bool {
			return strings.ToLower(field.Name) == "ignoredbyfunc"
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(args)
}
