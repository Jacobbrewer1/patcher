package inserter

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/jacobbrewer1/patcher"
)

func NewBatch(resources []any, opts ...BatchOpt) *SQLBatch {
	b := newBatchDefaults(opts...)

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
			if tag == patcher.TagOptSkip {
				continue
			}

			tags := strings.Split(tag, patcher.TagOptSeparator)
			if len(tags) > 1 {
				tag = tags[0]
			}

			// Skip unexported fields
			if !f.IsExported() {
				continue
			}

			// Skip fields that are to be ignored
			if b.checkSkipField(f) {
				continue
			}

			patcherOptsTag := f.Tag.Get(patcher.TagOptsName)
			if patcherOptsTag != "" {
				patcherOpts := strings.Split(patcherOptsTag, patcher.TagOptSeparator)
				if slices.Contains(patcherOpts, patcher.TagOptSkip) {
					continue
				}
			}

			// if no tag is set, use the field name
			if tag == "" {
				tag = f.Name
			}

			b.args = append(b.args, b.getFieldValue(v.Field(i), f))

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

func (b *SQLBatch) getFieldValue(v reflect.Value, f reflect.StructField) any {
	if f.Type.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}

	return v.Interface()
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

func (b *SQLBatch) checkSkipField(field reflect.StructField) bool {
	// The ignore fields tag takes precedence over the ignore fields list
	if b.checkSkipTag(field) {
		return true
	}

	// Check if the field is a primary key, we don't want to include the primary key in the insert unless specified
	if b.checkPrimaryKey(field) {
		return true
	}

	return b.ignoredFieldsCheck(field)
}

func (b *SQLBatch) checkSkipTag(field reflect.StructField) bool {
	val, ok := field.Tag.Lookup(patcher.TagOptsName)
	if !ok {
		return false
	}

	tags := strings.Split(val, patcher.TagOptSeparator)
	return slices.Contains(tags, patcher.TagOptSkip)
}

func (b *SQLBatch) checkPrimaryKey(field reflect.StructField) bool {
	// If we are including the primary key, we can immediately return false
	if b.includePrimaryKey {
		return false
	}

	val, ok := field.Tag.Lookup(patcher.DefaultDbTagName)
	if !ok {
		return false
	}

	tags := strings.Split(val, patcher.TagOptSeparator)
	return slices.Contains(tags, patcher.DBTagPrimaryKey)
}

func (b *SQLBatch) ignoredFieldsCheck(field reflect.StructField) bool {
	return b.checkIgnoredFields(strings.ToLower(field.Name)) || b.checkIgnoreFunc(field)
}

func (b *SQLBatch) checkIgnoreFunc(field reflect.StructField) bool {
	return b.ignoreFieldsFunc != nil && b.ignoreFieldsFunc(field)
}

func (b *SQLBatch) checkIgnoredFields(field string) bool {
	return len(b.ignoreFields) > 0 && slices.Contains(b.ignoreFields, strings.ToLower(field))
}
