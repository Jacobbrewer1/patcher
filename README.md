# Patcher

Patcher is a GO library that provides a simple way to generate and SQL patches from structs. The library was built out
of the need to generate patches for a database; when a new field is added to a struct, this would result in a bunch of
new `if` checks to be created in the codebase. This library aims to solve that problem by generating the SQL patches for
you.

## Usage

#### Basic

To use the library, you need to create a struct that represents the table you want to generate patches for. The struct
should have the following tags:

- `db:"column_name"`: This tag is used to specify the column name in the database.

Example:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
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
	const jsonStr = `{"name": "John", "age": 25}`

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
WHERE (1 = 1)
  AND (
    id = ?
    )
```

with the args:

```
["John", 25, 1]
```

#### Struct diffs

The Patcher library has functionality where you are able to inject changes from one struct to another. This is
configurable to include Zero values and Nil values if requested. Please see the
example [here](./examples/loader_with_opts) for the detailed example. Below is an example on how you can utilize this
method with the default behaviour (Please see the comment attached to the `LoadDiff` [method](./loader.go) for the
default behaviour).

Example:

```go
package main

import (
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Something struct {
	Number       int
	Text         string
	PrePopulated string
	NewText      string
}

func main() {
	s := Something{
		Number:       5,
		Text:         "Hello",
		PrePopulated: "PrePopulated",
	}

	n := Something{
		Number:  6,
		Text:    "Old Text",
		NewText: "New Text",
	}

	// The patcher.LoadDiff function will apply the changes from n to s.
	if err := patcher.LoadDiff(&s, &n); err != nil {
		panic(err)
	}

	// Output:
	// 6
	// Old Text
	// PrePopulated
	// New Text
	fmt.Println(s.Number)
	fmt.Println(s.Text)
	fmt.Println(s.PrePopulated)
	fmt.Println(s.NewText)
}

```

If you would like to generate an update script from two structs, you can use the `NewDiffSQLPatch` function. This
function will generate an update script from the two structs.

Example:

```go
package main

import (
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type Something struct {
	Number       int
	Text         string
	PrePopulated string
	NewText      string
}

type SomeWhere struct {
	id int
}

func NewSomeWhere(id int) *SomeWhere {
	return &SomeWhere{id: id}
}

func (s *SomeWhere) Where() (string, []any) {
	return "id = ?", []any{s.id}
}

func main() {
	s := Something{
		Number:       5,
		Text:         "Old Text",
		PrePopulated: "PrePopulated",
		NewText:      "New Text",
	}

	n := Something{
		Number:       5,
		Text:         "Old Text",
		PrePopulated: "PrePopulatedDifferent",
		NewText:      "New Text",
	}

	wherer := NewSomeWhere(5)

	// The patcher.LoadDiff function will apply the changes from n to s.
	patch, err := patcher.NewDiffSQLPatch(
		&s,
		&n,
		patcher.WithTable("table_name"),
		patcher.WithWhere(wherer),
	)
	if err != nil {
		panic(err)
	}

	sqlStr, sqlArgs, err := patch.GenerateSQL()
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlStr)
	fmt.Println(sqlArgs)
}

```

This will output:

```sql
UPDATE table_name
SET pre_populated = ?
WHERE (1 = 1)
  AND (
    id = ?
    )
```

with the args:

```
["PrePopulatedDifferent", 5]
```

You can also take a look at the Loader [examples](./examples) for more examples on how to use the library for this
approach.

#### Using `OR` in the where clause

If you would like to use `OR` in the where clause, you can apply the `patcher.WhereTyper` interface to your where
struct. Please take a look at the [example here](./examples/where_type).

### Joins

To generate a join, you need to create a struct that represents the join. This struct should implement
the [Joiner](./joiner.go) interface.

Once you have the join struct, you can pass it to the `GenerateSQL` function using the `WithJoin` option. You can add as
many of these as you would like.

# Examples

You can find examples of how to use this library in the [examples](./examples) directory.
