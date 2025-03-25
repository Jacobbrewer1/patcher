package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jacobbrewer1/patcher"
)

type Person struct {
	ID                *int    `db:"id" json:"id,omitempty"`
	Name              *string `db:"name" json:"name,omitempty"`
	IgnoredByTag      string  `db:"-" json:"ignored_by_tag,omitempty"`
	IncludedZeroValue string  `db:"includedZeroValue" json:"included_zero_value,omitempty"`
	IncludedNilValue  string  `db:"includedNilValue" json:"included_nil_value,omitempty"`
	IgnoredByFunc     string  `db:"ignoredFieldByFunc" json:"ignored_by_func,omitempty"`
	IgnoredByList     string  `db:"ignoredFieldByList" json:"ignored_by_list,omitempty"`
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
		patcher.WithIncludeZeroValues(true),
		patcher.WithIncludeNilValues(true),
		patcher.WithIgnoredFields("ignoredbylist"),
		patcher.WithIgnoredFieldsFunc(func(field *reflect.StructField) bool {
			return strings.EqualFold(field.Name, "ignoredbyfunc")
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(args)
}
