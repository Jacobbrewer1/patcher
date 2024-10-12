package inserter

import "database/sql"

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

func WithResources(resources []any) BatchOpt {
	return func(b *SQLBatch) {
		b.resources = resources
	}
}
