package inserter

import (
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/jacobbrewer1/patcher"
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
)

type SQLBatch struct {
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

	// ignoreFields is a list of fields to ignore when patching
	ignoreFields []string

	// ignoreFieldsFunc is a function that determines whether a field should be ignored
	//
	// This func should return true is the field is to be ignored
	ignoreFieldsFunc patcher.IgnoreFieldsFunc

	// includePrimaryKey determines whether the primary key should be included in the insert
	includePrimaryKey bool
}

// newBatchDefaults returns a new SQLBatch with default values
func newBatchDefaults(opts ...BatchOpt) *SQLBatch {
	b := &SQLBatch{
		fields:            make([]string, 0),
		args:              make([]any, 0),
		db:                nil,
		tagName:           patcher.DefaultDbTagName,
		table:             "",
		includePrimaryKey: false,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *SQLBatch) Fields() []string {
	if len(b.fields) == 0 {
		// Default behaviour to return nil if no fields are set
		return nil
	}
	return b.fields
}

func (b *SQLBatch) Args() []any {
	if len(b.args) == 0 {
		// Default behaviour to return nil if no args are set
		return nil
	}
	return b.args
}

func (b *SQLBatch) validateSQLGen() error {
	switch {
	case b.table == "":
		return ErrNoTable
	case len(b.fields) == 0:
		return ErrNoFields
	case len(b.args) == 0:
		return ErrNoArgs
	default:
		return nil
	}
}

func (b *SQLBatch) validateSQLInsert() error {
	switch {
	case b.db == nil:
		return ErrNoDatabaseConnection
	case b.table == "":
		return ErrNoTable
	case len(b.fields) == 0:
		return ErrNoFields
	case len(b.args) == 0:
		return ErrNoArgs
	default:
		return nil
	}
}

func (b *SQLBatch) checkSkipField(field *reflect.StructField) bool {
	return b.checkSkipTag(field) || b.checkPrimaryKey(field) || b.ignoredFieldsCheck(field)
}

func (b *SQLBatch) checkSkipTag(field *reflect.StructField) bool {
	val, ok := field.Tag.Lookup(patcher.TagOptsName)
	if !ok {
		return false
	}
	return slices.Contains(strings.Split(val, patcher.TagOptSeparator), patcher.TagOptSkip)
}

func (b *SQLBatch) checkPrimaryKey(field *reflect.StructField) bool {
	if b.includePrimaryKey {
		return false
	}
	val, ok := field.Tag.Lookup(patcher.DefaultDbTagName)
	if !ok {
		return false
	}
	return slices.Contains(strings.Split(val, patcher.TagOptSeparator), patcher.DBTagPrimaryKey)
}

func (b *SQLBatch) ignoredFieldsCheck(field *reflect.StructField) bool {
	return b.checkIgnoredFields(field.Name) || b.checkIgnoreFunc(field)
}

func (b *SQLBatch) checkIgnoreFunc(field *reflect.StructField) bool {
	return b.ignoreFieldsFunc != nil && b.ignoreFieldsFunc(field)
}

func (b *SQLBatch) checkIgnoredFields(field string) bool {
	return len(b.ignoreFields) > 0 && slices.Contains(b.ignoreFields, field)
}
