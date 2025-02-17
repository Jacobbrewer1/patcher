package inserter

import (
	"database/sql"
	"fmt"
	"reflect"
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
			f := t.Field(i)
			fVal := v.Field(i)

			if !patcher.IsValidType(fVal) || !f.IsExported() || b.checkSkipField(&f) {
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

			b.args = append(b.args, b.getFieldValue(fVal, &f))

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
	} else if f.Type.Kind() == reflect.Ptr {
		return v.Elem().Interface()
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

	placeholder := "(" + strings.Repeat("?, ", len(b.fields)-1) + "?)"
	placeholders := strings.Repeat(placeholder+", ", len(b.args)/len(b.fields))
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
