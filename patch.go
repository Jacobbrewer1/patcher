package patcher

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	// ErrNoDatabaseConnection is returned when no database connection is set
	ErrNoDatabaseConnection = errors.New("no database connection set")

	// ErrNoTable is returned when no table is set
	ErrNoTable = errors.New("no table set")

	// ErrNoFields is returned when no fields are set
	ErrNoFields = errors.New("no fields set")

	// ErrNoArgs is returned when no arguments are set
	ErrNoArgs = errors.New("no arguments set")

	// ErrNoWhere is returned when no where clause is set
	ErrNoWhere = errors.New("no where clause set")
)

type SQLPatch struct {
	fields []string
	args   []any

	// db is the database connection to use
	db *sql.DB

	// tagName is the tag name to look for in the struct. This is an override from the default tag "db"
	tagName string

	// table is the table name to use in the SQL statement
	table string

	// whereSql is the where clause to use in the SQL statement
	where strings.Builder

	// whereArgs is the arguments to use in the where clause
	whereArgs []any

	// joinSql is the join clause to use in the SQL statement
	joinSql strings.Builder

	// joinArgs is the arguments to use in the join clause
	joinArgs []any
}

func (s *SQLPatch) Fields() []string {
	return s.fields
}

func (s *SQLPatch) Args() []any {
	return s.args
}

func (s *SQLPatch) validatePerformPatch() error {
	if s.db == nil {
		return ErrNoDatabaseConnection
	} else if s.table == "" {
		return ErrNoTable
	} else if len(s.fields) == 0 {
		return ErrNoFields
	} else if len(s.args) == 0 {
		return ErrNoArgs
	} else if s.where.String() == "" {
		return ErrNoWhere
	}

	return nil
}

func (s *SQLPatch) validateSQLGen() error {
	if s.table == "" {
		return ErrNoTable
	} else if len(s.fields) == 0 {
		return ErrNoFields
	} else if len(s.args) == 0 {
		return ErrNoArgs
	} else if s.where.String() == "" {
		return ErrNoWhere
	}

	return nil
}
