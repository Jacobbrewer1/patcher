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
		t := reflect.TypeOf(r)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			continue
		}

		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		for i := range t.NumField() {
			if !patcher.IsValidType(v.Field(i)) {
				continue
			}

			f := t.Field(i)
			if !f.IsExported() || b.checkSkipField(&f) {
				continue
			}

			tag := f.Tag.Get(b.tagName)
			if tag == patcher.TagOptSkip {
				continue
			}

			if tag == "" {
				tag = f.Name
			} else {
				tag = strings.Split(tag, patcher.TagOptSeparator)[0]
			}

			b.args = append(b.args, b.getFieldValue(v.Field(i), &f))

			if _, ok := uniqueFields[tag]; ok {
				continue
			}

			b.fields = append(b.fields, tag)
			uniqueFields[tag] = struct{}{}
		}
	}
}

func (b *SQLBatch) getFieldValue(v reflect.Value, f *reflect.StructField) any {
	if f.Type.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	return v.Interface()
}

func (b *SQLBatch) GenerateSQL() (sqlStr string, args []any, err error) {
	if err := b.validateSQLGen(); err != nil {
		return "", nil, err
	}

	sqlBuilder := new(strings.Builder)
	sqlBuilder.WriteString("INSERT INTO ")
	sqlBuilder.WriteString(b.table)
	sqlBuilder.WriteString(" (")
	sqlBuilder.WriteString(strings.Join(b.fields, ", "))
	sqlBuilder.WriteString(") VALUES ")

	placeholder := strings.Repeat("?, ", len(b.fields))
	placeholder = "(" + placeholder[:len(placeholder)-2] + "), "
	placeholders := strings.Repeat(placeholder, len(b.args)/len(b.fields))
	sqlBuilder.WriteString(placeholders[:len(placeholders)-2])

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
