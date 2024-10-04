# Patcher

Patcher is a GO library that provides a simple way to generate and SQL patches from structs. The library was built out
of the need to generate patches for a database; when a new field is added to a struct, this would result in a bunch of
new `if` checks to be created in the codebase. This library aims to solve that problem by generating the SQL patches for
you.

## Usage

To use the library, you need to create a struct that represents the table you want to generate patches for. The struct
should have the following tags:

- `db:"column_name"`: This tag is used to specify the column name in the database.

Example:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/Jacobbrewer1/patcher"
)

type Person struct {
	ID   *int    `db:"id"`
	Name *string `db:"name"`
	Age  *int    `db:"age"`
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
	jsonStr := `{"name": "Jacob", "age": 25}`

	person := new(Person)
	if err := json.Unmarshal([]byte(jsonStr), person); err != nil {
		panic(err)
	}

	condition := NewPersonWhere(1)

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
```

This will output:

```sql
UPDATE people
SET name = ?,
    age  = ?
WHERE 1
  AND id = ?
```

with the args:

```
["Jacob", 25, 1]
```

### Joins

To generate a join, you need to create a struct that represents the join. This struct should implement
the [Joiner](./joiner.go) interface.

Once you have the join struct, you can pass it to the `GenerateSQL` function using the `WithJoin` option. You can add as
many of these as you would like.

# Examples

You can find examples of how to use this library in the [examples](./examples) directory.
