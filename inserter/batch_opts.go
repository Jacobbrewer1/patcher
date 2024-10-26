package inserter

import (
	"database/sql"

	"github.com/jacobbrewer1/patcher"
)

type BatchOpt func(*SQLBatch)

// WithTagName sets the tag name to look for in the struct. This is an override from the default tag "db"
func WithTagName(tagName string) BatchOpt {
	return func(b *SQLBatch) {
		b.tagName = tagName
	}
}

// WithTable sets the table name to use in the SQL statement
func WithTable(table string) BatchOpt {
	return func(b *SQLBatch) {
		b.table = table
	}
}

// WithDB sets the database connection to use
func WithDB(db *sql.DB) BatchOpt {
	return func(b *SQLBatch) {
		b.db = db
	}
}

// WithIgnoreFields sets the fields to ignore when patching
func WithIgnoreFields(fields ...string) BatchOpt {
	return func(b *SQLBatch) {
		b.ignoreFields = fields
	}
}

// WithIgnoreFieldsFunc sets the function that determines whether a field should be ignored
func WithIgnoreFieldsFunc(f patcher.IgnoreFieldsFunc) BatchOpt {
	return func(b *SQLBatch) {
		b.ignoreFieldsFunc = f
	}
}
