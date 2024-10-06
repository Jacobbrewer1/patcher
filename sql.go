package patcher

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const tagName = "db"

func NewSQLPatch(resource any, opts ...PatchOpt) *SQLPatch {
	sqlPatch := new(SQLPatch)
	sqlPatch.tagName = tagName

	for _, opt := range opts {
		opt(sqlPatch)
	}

	sqlPatch.patchGen(resource)

	return sqlPatch
}

func (s *SQLPatch) patchGen(resource any) {
	// If the resource is a pointer, we need to dereference it to get the value
	if reflect.TypeOf(resource).Kind() == reflect.Ptr {
		resource = reflect.ValueOf(resource).Elem().Interface()
	}

	// Ensure that the resource is a struct
	if reflect.TypeOf(resource).Kind() != reflect.Struct {
		// This is intentionally a panic as this is a programming error and should be fixed by the developer
		panic("resource is not a struct")
	}

	rType := reflect.TypeOf(resource)
	rVal := reflect.ValueOf(resource)
	n := rType.NumField()

	s.fields = make([]string, 0, n)
	s.args = make([]any, 0, n)

	for i := 0; i < n; i++ {
		fType := rType.Field(i)
		fVal := rVal.Field(i)
		tag := fType.Tag.Get(s.tagName)

		// skip nil properties (not going to be patched), skip unexported fields, skip fields to be skipped for SQL
		if fVal.Kind() == reflect.Ptr && (fVal.IsNil() || fType.PkgPath != "" || tag == "-") {
			continue
		} else if fVal.Kind() != reflect.Ptr && fVal.IsZero() {
			// skip zero values for non-pointer fields as we have no way to differentiate between zero values and nil pointers
			continue
		}

		// if no tag is set, use the field name
		if tag == "" {
			tag = fType.Name
		}
		// and make the tag lowercase in the end
		tag = strings.ToLower(tag)

		s.fields = append(s.fields, tag+" = ?")

		var val reflect.Value
		if fVal.Kind() == reflect.Ptr {
			val = fVal.Elem()
		} else {
			val = fVal
		}

		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s.args = append(s.args, val.Int())
		case reflect.String:
			s.args = append(s.args, val.String())
		case reflect.Bool:
			boolArg := 0
			if val.Bool() {
				boolArg = 1
			}
			s.args = append(s.args, boolArg)
		default:
			// This is intentionally a panic as this is a programming error and should be fixed by the developer
			panic("unhandled default case")
		}
	}
}

func GenerateSQL(resource any, opts ...PatchOpt) (string, []any, error) {
	sqlPatch := NewSQLPatch(resource, opts...)
	return sqlPatch.GenerateSQL()
}

func (s *SQLPatch) GenerateSQL() (string, []any, error) {
	if err := s.validateSQLGen(); err != nil {
		return "", nil, fmt.Errorf("validate perform patch: %w", err)
	}

	sqlBuilder := new(strings.Builder)
	sqlBuilder.WriteString("UPDATE ")
	sqlBuilder.WriteString(s.table)
	sqlBuilder.WriteString("\n")

	if s.joinSql.String() != "" {
		sqlBuilder.WriteString(s.joinSql.String())
	}

	sqlBuilder.WriteString("SET ")
	sqlBuilder.WriteString(strings.Join(s.fields, ", "))
	sqlBuilder.WriteString("\n")

	sqlBuilder.WriteString("WHERE 1\n")
	sqlBuilder.WriteString(s.where.String())

	args := append(s.joinArgs, s.args...)
	args = append(args, s.whereArgs...)

	return sqlBuilder.String(), args, nil
}

func PerformPatch(resource any, opts ...PatchOpt) (sql.Result, error) {
	sqlPatch := NewSQLPatch(resource, opts...)
	return sqlPatch.PerformPatch()
}

func (s *SQLPatch) PerformPatch() (sql.Result, error) {
	if err := s.validatePerformPatch(); err != nil {
		return nil, fmt.Errorf("validate perform patch: %w", err)
	}

	return s.db.Exec(s.GenerateSQL())
}
