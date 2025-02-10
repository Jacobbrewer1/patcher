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

type IgnoreFieldsFunc func(field *reflect.StructField) bool

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

// Fields returns the fields to update in the SQL statement
func (s *SQLPatch) Fields() []string {
	if len(s.fields) == 0 {
		// Default behaviour is to return nil if there are no fields
		return nil
	}
	return s.fields
}

// Args returns the arguments to use in the SQL statement
func (s *SQLPatch) Args() []any {
	if len(s.args) == 0 {
		// Default behaviour is to return nil if there are no args
		return nil
	}
	return s.args
}

// validatePerformPatch validates the SQLPatch for the PerformPatch method
func (s *SQLPatch) validatePerformPatch() error {
	switch {
	case s.db == nil:
		return ErrNoDatabaseConnection
	case s.table == "":
		return ErrNoTable
	case len(s.fields) == 0:
		return ErrNoFields
	case len(s.args) == 0:
		return ErrNoArgs
	case s.whereSql.String() == "":
		return ErrNoWhere
	default:
		return nil
	}
}

// validateSQLGen validates the SQLPatch for the SQLGen method
func (s *SQLPatch) validateSQLGen() error {
	switch {
	case s.table == "":
		return ErrNoTable
	case len(s.fields) == 0:
		return ErrNoFields
	case len(s.args) == 0:
		return ErrNoArgs
	case s.whereSql.String() == "":
		return ErrNoWhere
	default:
		return nil
	}
}

// shouldIncludeNil determines whether the field should be included in the patch
func (s *SQLPatch) shouldIncludeNil(tag string) bool {
	if s.includeNilValues {
		return true
	}

	return s.shouldOmitEmpty(tag)
}

// shouldIncludeZero determines whether zero values should be included in the patch
func (s *SQLPatch) shouldIncludeZero(tag string) bool {
	if s.includeZeroValues {
		return true
	}

	return s.shouldOmitEmpty(tag)
}

// shouldOmitEmpty determines whether the field should be omitted if it is empty
func (s *SQLPatch) shouldOmitEmpty(tag string) bool {
	if tag != "" {
		tags := strings.Split(tag, TagOptSeparator)
		if slices.Contains(tags, TagOptOmitempty) {
			return true
		}
	}

	return false
}

func (s *SQLPatch) shouldSkipField(fType *reflect.StructField, fVal reflect.Value) bool {
	if !fType.IsExported() || !isValidType(fVal) || s.checkSkipField(fType) {
		return true
	}

	patcherOptsTag := fType.Tag.Get(TagOptsName)
	if fVal.Kind() == reflect.Ptr && (fVal.IsNil() && !s.shouldIncludeNil(patcherOptsTag)) {
		return true
	}
	if fVal.Kind() != reflect.Ptr && (fVal.IsZero() && !s.shouldIncludeZero(patcherOptsTag)) {
		return true
	}
	if patcherOptsTag != "" {
		patcherOpts := strings.Split(patcherOptsTag, TagOptSeparator)
		if slices.Contains(patcherOpts, TagOptSkip) {
			return true
		}
	}
	return false
}

func (s *SQLPatch) checkSkipField(field *reflect.StructField) bool {
	// The ignore fields tag takes precedence over the ignore fields list
	if s.checkSkipTag(field) {
		return true
	}

	return s.ignoredFieldsCheck(field)
}

func (s *SQLPatch) checkSkipTag(field *reflect.StructField) bool {
	val, ok := field.Tag.Lookup(TagOptsName)
	if !ok {
		return false
	}

	tags := strings.Split(val, TagOptSeparator)
	return slices.Contains(tags, TagOptSkip)
}

func (s *SQLPatch) ignoredFieldsCheck(field *reflect.StructField) bool {
	return s.checkIgnoredFields(field.Name) || s.checkIgnoreFunc(field)
}

func (s *SQLPatch) checkIgnoreFunc(field *reflect.StructField) bool {
	return s.ignoreFieldsFunc != nil && s.ignoreFieldsFunc(field)
}

func (s *SQLPatch) checkIgnoredFields(field string) bool {
	return len(s.ignoreFields) > 0 && slices.Contains(s.ignoreFields, field)
}
