package patcher

import (
	"database/sql"
)

const (
	TagOptsName     = "patcher"
	TagOptSeparator = ","
	TagOptSkip      = "-"
	TagOptOmitempty = "omitempty"
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

// WithFilter takes in either a Wherer or a Joiner to set the filter to use in the SQL statement
func WithFilter(filter any) PatchOpt {
	return func(s *SQLPatch) {
		if join, ok := filter.(Joiner); ok {
			WithJoin(join)(s)
		}

		if where, ok := filter.(Wherer); ok {
			WithWhere(where)(s)
		}
	}
}

// WithWhere sets the where clause to use in the SQL statement
func WithWhere(where Wherer) PatchOpt {
	return func(s *SQLPatch) {
		appendWhere(where, s.whereSql, &s.whereArgs)
	}
}

// WithWhereStr takes a string and args to set the where clause to use in the SQL statement. This is useful when you
// want to use a simple where clause.
//
// Note. The where string should not contain the "WHERE" keyword. We recommend using the WhereTyper interface if you
// want to specify the WHERE type or do a more complex WHERE clause.
func WithWhereStr(where string, args ...any) PatchOpt {
	return func(s *SQLPatch) {
		appendWhere(&whereStringOption{
			where: where,
			args:  args,
		}, s.whereSql, &s.whereArgs)
	}
}

// WithJoin sets the join clause to use in the SQL statement
func WithJoin(join Joiner) PatchOpt {
	return func(s *SQLPatch) {
		appendJoin(join, s.joinSql, &s.joinArgs)
	}
}

// WithJoinStr takes a string and args to set the join clause to use in the SQL statement. This is useful when you
// want to use a simple join clause.
//
// Note. The join string should not contain the "JOIN" keyword. We recommend using the Joiner interface if you
// want to specify the JOIN type or do a more complex JOIN clause.
func WithJoinStr(join string, args ...any) PatchOpt {
	return func(s *SQLPatch) {
		appendJoin(&joinStringOption{
			join: join,
			args: args,
		}, s.joinSql, &s.joinArgs)
	}
}

// WithDB sets the database connection to use
func WithDB(db *sql.DB) PatchOpt {
	return func(s *SQLPatch) {
		s.db = db
	}
}

// WithIncludeZeroValues sets whether zero values should be included in the patch.
//
// This is useful when you want to set a field to zero.
func WithIncludeZeroValues(includeZeroValues bool) PatchOpt {
	return func(s *SQLPatch) {
		s.includeZeroValues = includeZeroValues
	}
}

// WithIncludeNilValues sets whether nil values should be included in the patch.
//
// This is useful when you want to set a field to nil.
func WithIncludeNilValues(includeNilValues bool) PatchOpt {
	return func(s *SQLPatch) {
		s.includeNilValues = includeNilValues
	}
}

// WithIgnoredFields sets the fields to ignore when patching.
//
// This should be the actual field name, not the JSON tag name or the db tag name.
//
// Note. When we parse the slice of strings, we convert them to lowercase to ensure that the comparison is
// case-sensitive.
func WithIgnoredFields(fields ...string) PatchOpt {
	return func(s *SQLPatch) {
		s.ignoreFields = fields
	}
}

// WithIgnoredFieldsFunc sets a function that determines whether a field should be ignored when patching.
func WithIgnoredFieldsFunc(f IgnoreFieldsFunc) PatchOpt {
	return func(s *SQLPatch) {
		s.ignoreFieldsFunc = f
	}
}

// WithLimit sets the limit for the SQL query.
func WithLimit(limit int) PatchOpt {
	return func(s *SQLPatch) {
		s.limit = limit
	}
}

// WithOffset sets the offset for the SQL query.
func WithOffset(offset int) PatchOpt {
	return func(s *SQLPatch) {
		s.offset = offset
	}
}
