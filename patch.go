package patcher

import (
	"database/sql"
	"errors"
	"reflect"
	"slices"
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

type IgnoreFieldsFunc func(field reflect.StructField) bool

type SQLPatch struct {
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

	// whereSql is the where clause to use in the SQL statement
	whereSql *strings.Builder

	// whereArgs is the arguments to use in the where clause
	whereArgs []any

	// joinSql is the join clause to use in the SQL statement
	joinSql *strings.Builder

	// joinArgs is the arguments to use in the join clause
	joinArgs []any

	// includeZeroValues determines whether zero values should be included in the patch
	includeZeroValues bool

	// includeNilValues determines whether nil values should be included in the patch
	includeNilValues bool

	// ignoreFields is a list of fields to ignore when patching
	ignoreFields []string

	// ignoreFieldsFunc is a function that determines whether a field should be ignored
	//
	// This func should return true is the field is to be ignored
	ignoreFieldsFunc IgnoreFieldsFunc
}

// newPatchDefaults creates a new SQLPatch with default options.
func newPatchDefaults(opts ...PatchOpt) *SQLPatch {
	// Default options
	p := &SQLPatch{
		fields:            make([]string, 0),
		args:              make([]any, 0),
		db:                nil,
		tagName:           DefaultDbTagName,
		table:             "",
		whereSql:          new(strings.Builder),
		whereArgs:         nil,
		joinSql:           new(strings.Builder),
		joinArgs:          nil,
		includeZeroValues: false,
		includeNilValues:  false,
		ignoreFields:      nil,
		ignoreFieldsFunc:  nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
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
	} else if s.whereSql.String() == "" {
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
	} else if s.whereSql.String() == "" {
		return ErrNoWhere
	}

	return nil
}

func (s *SQLPatch) shouldIncludeNil(tag string) bool {
	if s.includeNilValues {
		return true
	}

	if tag != "" {
		tags := strings.Split(tag, TagOptSeparator)
		if slices.Contains(tags, TagOptAllowNil) {
			return true
		}
	}

	return false
}

func (s *SQLPatch) shouldIncludeZero(tag string) bool {
	if s.includeZeroValues {
		return true
	}

	if tag != "" {
		tagOpts := strings.Split(tag, TagOptSeparator)
		if slices.Contains(tagOpts, TagOptAllowZero) {
			return true
		}
	}

	return false
}
