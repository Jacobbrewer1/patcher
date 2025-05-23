# Patcher

[![Go Reference](https://pkg.go.dev/badge/github.com/jacobbrewer1/patcher.svg)](https://pkg.go.dev/github.com/jacobbrewer1/patcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/jacobbrewer1/patcher)](https://goreportcard.com/report/github.com/jacobbrewer1/patcher)

Patcher is a GO library that provides a simple way to generate and SQL patches from structs. The library was built out
of the need to generate patches for a database; when a new field is added to a struct, this would result in a bunch of
new `if` checks to be created in the codebase. This library aims to solve that problem by generating the SQL patches for
you.

## What is Patcher?

* **Automatic SQL Generation**: It automatically generates SQL UPDATE queries from structs, reducing the need for
  manually
  writing and maintaining SQL statements.
* **Code Simplification**: It reduces the amount of boilerplate code and if-else conditions required to handle different
  struct fields, making the codebase cleaner and easier to maintain.
* **Struct Diffs**: It allows injecting changes from one struct to another and generating update scripts based on
  differences, streamlining the process of synchronizing data changes.
* **Join Support**: It supports generating SQL joins by creating structs that implement the Joiner interface,
  simplifying the process of managing related data across multiple tables.

## Why Use Patcher?

* **Saves Time**: It saves time by automatically generating SQL queries from structs, reducing the need to write and
  maintain SQL statements manually.
* **Reduces Errors**: It reduces the risk of errors by automatically generating SQL queries based on struct fields,
  eliminating the need to manually update queries when struct fields change.
* **Simplifies Code**: It simplifies the codebase by reducing the amount of boilerplate code and if-else conditions
  required to handle different struct fields, making the code easier to read and maintain.
* **Streamlines Data Synchronization**: It streamlines the process of synchronizing data changes by allowing you to
  inject changes from one struct to another and generate update scripts based on differences.
* **Supports Joins**: It supports generating SQL joins by creating structs that implement the Joiner interface, making
  it easier to manage related data across multiple tables.
* **Flexible Configuration**: It provides flexible configuration options to customize the SQL generation process, such
  as including zero or nil values in the diff.
* **Easy Integration**: It is easy to integrate into existing projects and can be used with any Go project that needs to
  generate SQL queries from structs.
* **Open Source**: It is open-source and available under the Apache 2.0 license.

### Demonstration

Here is a simple example of how Patcher would compare to an example of not using Patcher.

#### Without Patcher

```go
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	ID    int     `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func updateUser(db *sql.DB, user User) error {
	query := "UPDATE users SET"
	var updates []string
	var args []any

	if user.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *user.Name)
	}
	if user.Email != nil {
		updates = append(updates, "email = ?")
		args = append(args, *user.Email)
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	query += " " + strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, user.ID)

	_, err := db.Exec(query, args...)
	return err
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := User{ID: userID}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateUser(db, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", patchHandler).Methods("PATCH")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

```

As demonstrated above, without Patcher, you would need to manually write the SQL update query and handle each field. If
a new field is added to the struct, you would need to update the query manually, which can be error-prone (e.g., missing
or forgetting to update a field) and time-consuming.

#### With Patcher

```go
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jacobbrewer1/patcher"
)

type User struct {
	ID    *int    `db:"id" json:"id,omitempty"`
	Name  *string `db:"name" json:"name,omitempty"`
	Email *string `db:"email" json:"email,omitempty"`
}

type UserWhere struct {
	ID *int `db:"id"`
}

func NewUserWhere(id int) *UserWhere {
	return &UserWhere{ID: &id}
}

func (u *UserWhere) Where() (sqlStr string, sqlArgs []any) {
	return "id = ?", []any{*u.ID}
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)
	idStr := vars["id"]
	parsedID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := new(User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.ID = &parsedID
	condition := NewUserWhere(parsedID)

	sqlStr, args, err := patcher.GenerateSQL(
		user,
		patcher.WithTable("users"),
		patcher.WithWhere(condition),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, execErr := db.Exec(sqlStr, args...)
	if execErr != nil {
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", patchHandler).Methods("PATCH")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

```

With Patcher, you can generate the SQL update query automatically from the struct fields, reducing the need to write and
maintain the query manually. If a new field is added to the struct, Patcher will automatically include it in the update
query, simplifying the code and reducing the risk of errors.

## Usage

### Configuration

#### LoadDiff Options

* `includeZeroValues`: Set to true to include zero values in the diff.
* `includeNilValues`: Set to true to include nil values in the diff.

#### GenerateSQL Options

* `WithTable(tableName string)`: Specify the table name for the SQL query.
* `WithWhere(whereClause Wherer)`: Provide a where clause for the SQL query.
    * You can pass a struct that implements the `WhereTyper` interface to use `OR` in the where clause. Patcher will
      default to `AND` if the `WhereTyper` interface is not implemented.
* `WithJoin(joinClause Joiner)`: Add join clauses to the SQL query.
* `includeZeroValues`: Set to true to include zero values in the Patch.
* `includeNilValues`: Set to true to include nil values in the Patch.

### Basic Examples

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
	ID   *int    `db:"-"`
	Name *string `db:"name"`
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

	fmt.Println(sqlStr)

	fmt.Println(args)
}

```

This will output:

```sql
UPDATE people
SET name = ?
WHERE (1 = 1)
  AND (
    id = ?
    )
```

with the args:

```
["john", 1]
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

	fmt.Println(s.Number)
	fmt.Println(s.Text)
	fmt.Println(s.PrePopulated)
	fmt.Println(s.NewText)
}

```

This will output:

```
6
Hello
PrePopulated
New Text
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

## Installation

To install the Patcher library, use the following command:

```sh
go get github.com/jacobbrewer1/patcher
```

## Examples

You can find examples of how to use this library in the [examples](./examples) directory.

## Contributing

We welcome contributions! Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Write tests for your changes.
4. Run the tests to ensure everything works.
5. Submit a pull request.

To run tests, use the following command:

```sh
go test ./...
```
