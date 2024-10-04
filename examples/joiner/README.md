# Joiner

For this example assume that you have two tables in the database `people` and `contacts`.

## People

| id | name | age |
|----|------|-----|
| 1  | Ben  | 25  |
| 2  | John | 30  |

## Contacts

| id | person_id | email               |
|----|-----------|---------------------|
| 1  | 1         | ben@example.com     |
| 2  | 2         | john@exampletwo.com |

# Example

Moving forwards with the example, lets assume that you are updating the person and you have received their email as the
key on the data. You can use the `Joiner` interface to join the two tables and update the person.

```go
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
	ID *int
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

type PersonContactJoiner struct{}

func NewPersonContactJoiner() *PersonContactJoiner {
	return &PersonContactJoiner{}
}

func (p *PersonContactJoiner) Join() (string, []any) {
	return "JOIN contacts c ON c.person_id = p.id", nil
}

func main() {
	jsonStr := `{"name": "john", "age": 25, "email": "john@exampletwo.com"}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(1)
	joiner := NewPersonContactJoiner()

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

```

```go
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
	jsonStr := `{"name": "john", "age": 25, "email": "john@exampletwo.com"}`

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

```
