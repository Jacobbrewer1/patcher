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

func NewBatch(resources []any, opts ...BatchOpt) *SQLBatch {
	b := new(SQLBatch)
	b.tagName = defaultTagName
	for _, opt := range opts {
		opt(b)
	}

	b.genBatch(resources)

	return b
}

func (b *SQLBatch) genBatch(resources []any) {
	uniqueFields := make(map[string]struct{})

	for _, r := range resources {
		// get the type of the resource
		t := reflect.TypeOf(r)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// Is the type a struct?
		if t.Kind() != reflect.Struct {
			continue
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

func (b *SQLBatch) GenerateSQL() (string, []any, error) {
	if err := b.validateSQLGen(); err != nil {
		return "", nil, err
	}

	sqlBuilder := new(strings.Builder)

	sqlBuilder.WriteString("INSERT INTO ")
	sqlBuilder.WriteString(b.table)
	sqlBuilder.WriteString(" (")
	sqlBuilder.WriteString(strings.Join(b.fields, ", "))
	sqlBuilder.WriteString(") VALUES ")

	// We need to have the same number of "?" as fields and then repeat that for the number of resources
	placeholder := strings.Repeat("?, ", len(b.fields))
	placeholder = placeholder[:len(placeholder)-2] // Remove the trailing ", "
	placeholder = "(" + placeholder + "), "

	// Calculate the number of placeholders needed. Args divided by fields
	n := len(b.args) / len(b.fields)

	// Repeat the placeholder for the number of resources
	placeholders := strings.Repeat(placeholder, n)
	sqlBuilder.WriteString(placeholders[:len(placeholders)-2]) // Remove the trailing ", " and add the closing ")"

	return sqlBuilder.String(), b.args, nil
}

func (b *SQLBatch) Perform() (sql.Result, error) {
	if err := b.validateSQLInsert(); err != nil {
		return nil, fmt.Errorf("validate SQL generation: %w", err)
	}

	sqlStr, args, err := b.GenerateSQL()
	if err != nil {
		return nil, fmt.Errorf("generate SQL: %w", err)
	}

	return b.db.Exec(sqlStr, args...)
}
