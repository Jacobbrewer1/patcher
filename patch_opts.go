package patcher

import (
	"database/sql"
	"strings"
)

const (
	TagOptsName     = "patcher"
	TagOptSeparator = ","
	TagOptSkip      = "-"
	TagOptAllowNil  = "nil"
	TagOptAllowZero = "zero"
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
		if s.whereSql == nil {
			s.whereSql = new(strings.Builder)
		}
		fwSQL, fwArgs := where.Where()
		if fwArgs == nil {
			fwArgs = make([]any, 0)
		}
		wtStr := WhereTypeAnd // default to AND
		wt, ok := where.(WhereTyper)
		if ok && wt.WhereType().IsValid() {
			wtStr = wt.WhereType()
		}
		s.whereSql.WriteString(string(wtStr) + " ")
		s.whereSql.WriteString(strings.TrimSpace(fwSQL))
		s.whereSql.WriteString("\n")
		s.whereArgs = append(s.whereArgs, fwArgs...)
	}
}

// WithJoin sets the join clause to use in the SQL statement
func WithJoin(join Joiner) PatchOpt {
	return func(s *SQLPatch) {
		if s.joinSql == nil {
			s.joinSql = new(strings.Builder)
		}
		fjSQL, fjArgs := join.Join()
		if fjArgs == nil {
			fjArgs = make([]any, 0)
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

// WithIncludeZeroValues sets whether zero values should be included in the patch.
//
// This is useful when you want to set a field to zero.
func WithIncludeZeroValues() PatchOpt {
	return func(s *SQLPatch) {
		s.includeZeroValues = true
	}
}

// WithIncludeNilValues sets whether nil values should be included in the patch.
//
// This is useful when you want to set a field to nil.
func WithIncludeNilValues() PatchOpt {
	return func(s *SQLPatch) {
		s.includeNilValues = true
	}
}

// WithIgnoredFields sets the fields to ignore when patching.
//
// This should be the actual field name, not the JSON tag name or the db tag name.
//
// Note. When we parse the slice of strings, we convert them to lowercase to ensure that the comparison is
// case-insensitive.
func WithIgnoredFields(fields ...string) PatchOpt {
	return func(s *SQLPatch) {
		for i := range fields {
			fields[i] = strings.ToLower(fields[i])
		}

		s.ignoreFields = fields
	}
}

// WithIgnoredFieldsFunc sets a function that determines whether a field should be ignored when patching.
func WithIgnoredFieldsFunc(f IgnoreFieldsFunc) PatchOpt {
	return func(s *SQLPatch) {
		s.ignoreFieldsFunc = f
	}
}
