# Inserter Package

The `inserter` package provides functionality to insert data into a database using Go. It is designed to be flexible and easy to use.

## Installation

To install the `inserter` package, use the following command:

```sh
go get github.com/jacobbrewer1/patcher/inserter
```

## Usage

Here is an example of how to use the inserter package:

```go
package main

import (
	"fmt"

	"github.com/jacobbrewer1/patcher/inserter"
)

type User struct {
	ID    int    `db:"id,pk,autoinc"` // pk = primary key (This field will be ignored by default by the inserter package), autoinc = auto increment
	Name  string `db:"name"`
	Email string `db:"email"`
}

func main() {
	user := User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	sql, args, err := inserter.NewBatch([]any{user}, inserter.WithTable("users")).GenerateSQL()
	if err != nil {
		panic(err)
	}

	fmt.Println(sql)
	fmt.Println(args)
}

```

This will output the following:

```SQL
INSERT INTO users (id, name, email) VALUES (?, ?, ?)
```

with the following arguments:

```
[1, "John Doe", "john.doe@example.com"]
```

## Configuration Options

### GenerateInsertSQL Options

* `WithTable(tableName string)`: Specify the table name for the SQL query.

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

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.
