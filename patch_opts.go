package patcher

import (
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
)

type PatchOpt func(*SQLPatch)

// WithTagName sets the tag name to look for in the struct. This is an override from the default tag "db"
func WithTagName(tagName string) PatchOpt {
	return func(s *SQLPatch) {
		s.tagName = tagName
	}
}

// WithTable sets the table name to use in the SQL statement
func WithTable(table string) PatchOpt {
	return func(s *SQLPatch) {
		s.table = table
	}
}

// WithWhere sets the where clause to use in the SQL statement
func WithWhere(where Wherer) PatchOpt {
	return func(s *SQLPatch) {
		fwSQL, fwArgs := where.Where()
		if fwArgs == nil {
			fwArgs = []any{}
		}
		s.where.WriteString("AND ")
		s.where.WriteString(strings.TrimSpace(fwSQL))
		s.where.WriteString("\n")
		s.whereArgs = append(s.whereArgs, fwArgs...)
	}
}

// WithJoin sets the join clause to use in the SQL statement
func WithJoin(join Joiner) PatchOpt {
	return func(s *SQLPatch) {
		fjSQL, fjArgs := join.Join()
		if fjArgs == nil {
			fjArgs = []any{}
		}
		s.joinSql.WriteString(strings.TrimSpace(fjSQL))
		s.joinSql.WriteString("\n")
		s.joinArgs = append(s.joinArgs, fjArgs...)
	}
}

// WithDB sets the database connection to use
func WithDB(db *sql.DB) PatchOpt {
	return func(s *SQLPatch) {
		s.db = db
	}
}

// WithSQLxDB sets the database from an SQLx connection
func WithSQLxDB(db *sqlx.DB) PatchOpt {
	return func(s *SQLPatch) {
		s.db = db.DB
	}
}
