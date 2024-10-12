package inserter

import (
	"database/sql"
	"errors"
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

	// ErrNoResources is returned when no resources are set
	ErrNoResources = errors.New("no resources set")
)

type SQLBatch struct {
	// resources is the resources to use in the SQL statement
	resources []any

	// fields is the fields to update in the SQL statement
	fields []string

	// args is the arguments to use in the SQL statement
	args []any

	// db is the database connection to use
	db *sql.DB

	// tagName is the tag name to look for in the struct. This is an override from the default tag "db"
	tagName string

	// table is the table name to use in the SQL statement
	table string
}

func (b *SQLBatch) AddResources(resources ...any) {
	b.resources = append(b.resources, resources...)
}

func (b *SQLBatch) Fields() []string {
	return b.fields
}

func (b *SQLBatch) Args() []any {
	return b.args
}

func (b *SQLBatch) validateSQLGen() error {
	if b.table == "" {
		return ErrNoTable
	}
	if len(b.resources) == 0 {
		return ErrNoResources
	}
	return nil
}

func (b *SQLBatch) validateSQLInsert() error {
	if b.db == nil {
		return ErrNoDatabaseConnection
	}
	if b.table == "" {
		return ErrNoTable
	}
	if len(b.resources) == 0 {
		return ErrNoResources
	}
	return nil
}
