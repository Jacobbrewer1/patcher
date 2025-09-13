package main

import (
	"encoding/json"
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type User struct {
	ID    *int    `db:"id" patcher:"-" json:"id,omitempty"`
	Name  *string `db:"name" json:"name,omitempty"`
	Email *string `db:"email" json:"email,omitempty"`
}

type UserWhere struct {
	ID *int `db:"id"`
}

func NewUserWhere(id int) *UserWhere {
	return &UserWhere{
		ID: &id,
	}
}

func (u *UserWhere) Where() (sqlStr string, sqlArgs []any) {
	return "id = ?", []any{*u.ID}
}

func main() {
	const jsonStr = `{"id": 1, "name": "john", "email": "john@example.com"}`

	user := new(User)
	if err := json.Unmarshal([]byte(jsonStr), user); err != nil {
		panic(err)
	}

	condition := NewUserWhere(*user.ID)

	// Generate SQL for PostgreSQL (using $1, $2, $3 placeholders)
	sqlStr, args, err := patcher.GenerateSQL(
		user,
		patcher.WithTable("users"),
		patcher.WithWhere(condition),
		patcher.WithDialect(patcher.DialectPostgreSQL),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("PostgreSQL SQL:")
	fmt.Println(sqlStr)
	fmt.Println("Args:")
	fmt.Println(args)
	fmt.Println()

	// For comparison, generate MySQL/SQLite SQL (using ? placeholders)
	sqlStrMySQL, argsMySQL, err := patcher.GenerateSQL(
		user,
		patcher.WithTable("users"),
		patcher.WithWhere(condition),
		patcher.WithDialect(patcher.DialectMySQL),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("MySQL/SQLite SQL:")
	fmt.Println(sqlStrMySQL)
	fmt.Println("Args:")
	fmt.Println(argsMySQL)
}