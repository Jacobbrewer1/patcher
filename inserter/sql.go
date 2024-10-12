package inserter

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const (
	// defaultTagName is the default tag name to look for in the struct
	defaultTagName = "db"
)

func NewBatch(opts ...BatchOpt) *SQLBatch {
	b := new(SQLBatch)
	b.tagName = defaultTagName
	for _, opt := range opts {
		opt(b)
	}

	b.genBatch()

	return b
}

func (b *SQLBatch) genBatch() {
	uniqueFields := make(map[string]struct{})

	for _, r := range b.resources {
		// get the type of the resource
		t := reflect.TypeOf(r)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// get the value of the resource
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// get the fields
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get(b.tagName)
			if tag == "-" {
				continue
			}

			if !f.IsExported() {
				continue
			}

			// if no tag is set, use the field name
			if tag == "" {
				tag = strings.ToLower(f.Name)
			}

			b.args = append(b.args, v.Field(i).Interface())

			// if the field is not unique, skip it
			if _, ok := uniqueFields[tag]; ok {
				continue
			}

			// add the field to the list
			b.fields = append(b.fields, tag)
			uniqueFields[tag] = struct{}{}
		}
	}
}

func (b *SQLBatch) sqlGen() (string, []any, error) {
	if err := b.validateSQLGen(); err != nil {
		return "", nil, err
	}

	sqlBuilder := new(strings.Builder)

	sqlBuilder.WriteString("INSERT INTO ")
	sqlBuilder.WriteString(b.table)
	sqlBuilder.WriteString(" (")
	sqlBuilder.WriteString(strings.Join(b.fields, ", "))
	sqlBuilder.WriteString(") VALUES ")

	// Repeat "?" for the number of fields separated by ", "
	sqlBuilder.WriteString("(")

	placeholders := strings.Repeat("?, ", len(b.args))
	sqlBuilder.WriteString(placeholders[:len(placeholders)-2]) // Remove the trailing ", " and add the closing ")"
	sqlBuilder.WriteString(")")

	return sqlBuilder.String(), b.args, nil
}

func (b *SQLBatch) Perform() (sql.Result, error) {

	sqlStr, args, err := b.sqlGen()
	if err != nil {
		return nil, fmt.Errorf("generate SQL: %w", err)
	}

	return b.db.Exec(sqlStr, args...)
}
