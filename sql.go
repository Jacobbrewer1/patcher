package patcher

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	DefaultDbTagName = "db"
	DBTagPrimaryKey  = "pk"
)

var (
	// ErrNoChanges is returned when no changes are detected between the old and new objects
	ErrNoChanges = errors.New("no changes detected between the old and new objects")
)

// NewSQLPatch creates a new SQLPatch instance with the given resource and options.
// It initializes the SQLPatch with default settings and generates the SQL patch
// for the provided resource by processing its fields and applying the necessary tags and options.
func NewSQLPatch(resource any, opts ...PatchOpt) *SQLPatch {
	sqlPatch := newPatchDefaults(opts...)
	sqlPatch.patchGen(resource)
	return sqlPatch
}

// patchGen generates the SQL patch for the given resource.
// It processes the fields of the struct, applying the necessary tags and options,
// and prepares the SQL update statement components (fields and arguments).
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

		// Skip unexported fields
		if !fType.IsExported() {
			continue
		}

		tags := strings.Split(tag, TagOptSeparator)
		if len(tags) > 1 {
			tag = tags[0]
		}

		patcherOptsTag := fType.Tag.Get(TagOptsName)

		// Skip fields that are to be ignored
		if s.checkSkipField(fType) {
			continue
		} else if fVal.Kind() == reflect.Ptr && (fVal.IsNil() && !s.shouldIncludeNil(patcherOptsTag)) {
			continue
		} else if fVal.Kind() != reflect.Ptr && (fVal.IsZero() && !s.shouldIncludeZero(patcherOptsTag)) {
			continue
		}

		if patcherOptsTag != "" {
			patcherOpts := strings.Split(patcherOptsTag, TagOptSeparator)
			if slices.Contains(patcherOpts, TagOptSkip) {
				continue
			}
		}

		// If no tag is set, use the field name
		if tag == "" {
			tag = fType.Name
		}

		addField := func() {
			s.fields = append(s.fields, tag+" = ?")
		}

		if fVal.Kind() == reflect.Ptr && fVal.IsNil() && s.shouldIncludeNil(patcherOptsTag) {
			s.args = append(s.args, nil)
			addField()
			continue
		} else if fVal.Kind() == reflect.Ptr && fVal.IsNil() {
			continue
		}

		addField()

		val := fVal
		if fVal.Kind() == reflect.Ptr {
			val = fVal.Elem()
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
		case reflect.Float32, reflect.Float64:
			s.args = append(s.args, val.Float())
		default:
			// This is intentionally a panic as this is a programming error and should be fixed by the developer
			panic(fmt.Sprintf("unsupported type: %s", val.Kind()))
		}
	}
}

// GenerateSQL generates the SQL update statement and its arguments for the given resource.
// It creates a new SQLPatch instance with the provided options, processes the resource's fields,
// and constructs the SQL update statement along with the necessary arguments.
func GenerateSQL(resource any, opts ...PatchOpt) (string, []any, error) {
	return NewSQLPatch(resource, opts...).GenerateSQL()
}

// GenerateSQL constructs the SQL update statement and its arguments.
// It validates the SQL generation process, builds the SQL update statement
// with the table name, join clauses, set clauses, and where clauses,
// and returns the final SQL string along with the arguments.
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

	sqlBuilder.WriteString("WHERE (1=1)\n")
	sqlBuilder.WriteString("AND (\n")

	// If the where clause starts with "AND" or "OR", we need to remove it
	where := s.whereSql.String()
	if strings.HasPrefix(where, string(WhereTypeAnd)) || strings.HasPrefix(where, string(WhereTypeOr)) {
		where = strings.TrimPrefix(where, string(WhereTypeAnd))
		where = strings.TrimPrefix(where, string(WhereTypeOr))
		where = strings.TrimSpace(where)
	}

	sqlBuilder.WriteString(strings.TrimSpace(where) + "\n")
	sqlBuilder.WriteString(")")

	args := append(s.joinArgs, s.args...)
	args = append(args, s.whereArgs...)

	return sqlBuilder.String(), args, nil
}

// PerformPatch executes the SQL update statement for the given resource.
// It creates a new SQLPatch instance with the provided options, generates the SQL update statement,
// and executes it using the database connection.
func PerformPatch(resource any, opts ...PatchOpt) (sql.Result, error) {
	return NewSQLPatch(resource, opts...).PerformPatch()
}

// PerformDiffPatch executes the SQL update statement for the differences between the old and new resources.
// It creates a new SQLPatch instance by comparing the old and new resources, generates the SQL update statement,
// and executes it using the database connection.
func PerformDiffPatch[T any](old, newT *T, opts ...PatchOpt) (sql.Result, error) {
	sqlPatch, err := NewDiffSQLPatch(old, newT, opts...)
	if err != nil {
		return nil, fmt.Errorf("new diff sql patch: %w", err)
	}

	return sqlPatch.PerformPatch()
}

// PerformPatch executes the SQL update statement for the current SQLPatch instance.
// It validates the SQL generation process, constructs the SQL update statement and its arguments,
// and executes the statement using the database connection.
// It returns the result of the SQL execution or an error if the process fails.
func (s *SQLPatch) PerformPatch() (sql.Result, error) {
	if err := s.validatePerformPatch(); err != nil {
		return nil, fmt.Errorf("validate perform patch: %w", err)
	}

	sqlStr, args, err := s.GenerateSQL()
	if err != nil {
		return nil, fmt.Errorf("generate SQL: %w", err)
	}

	return s.db.Exec(sqlStr, args...)
}

// NewDiffSQLPatch creates a new SQLPatch instance by comparing the old and new resources.
// It initializes the SQLPatch with default settings, loads the differences between the old and new resources,
// and prepares the SQL update statement components (fields and arguments) for the differences.
func NewDiffSQLPatch[T any](old, newT *T, opts ...PatchOpt) (*SQLPatch, error) {
	if !isPointerToStruct(old) || !isPointerToStruct(newT) {
		return nil, ErrInvalidType
	}

	// Take a copy of the old object
	oldCopy := reflect.New(reflect.TypeOf(old).Elem()).Interface()

	// copy the old object into the copy
	reflect.ValueOf(oldCopy).Elem().Set(reflect.ValueOf(old).Elem())

	patch := newPatchDefaults(opts...)
	if err := patch.loadDiff(old, newT); err != nil {
		return nil, fmt.Errorf("load diff: %w", err)
	}

	// Are the old and new objects the same?
	if reflect.DeepEqual(old, oldCopy) {
		return nil, ErrNoChanges
	}

	oldElem := reflect.ValueOf(old).Elem()
	oldCopyElem := reflect.ValueOf(oldCopy).Elem()

	// For each field in the old object, compare it against the copy and if the fields are the same, set them to zero or nil.
	for i := 0; i < reflect.ValueOf(old).Elem().NumField(); i++ {
		oldField := oldElem.Field(i)
		copyField := oldCopyElem.Field(i)

		patcherOptsTag := oldElem.Type().Field(i).Tag.Get(TagOptsName)

		if oldField.Kind() == reflect.Ptr && (oldField.IsNil() && copyField.IsNil() && !patch.shouldIncludeNil(patcherOptsTag)) {
			continue
		} else if oldField.Kind() != reflect.Ptr && (oldField.IsZero() && copyField.IsZero() && !patch.shouldIncludeZero(patcherOptsTag)) {
			continue
		}

		if reflect.DeepEqual(oldField.Interface(), copyField.Interface()) {
			// Field is the same, set it to zero or nil. Add it to be ignored in the patch
			if patch.ignoreFields == nil {
				patch.ignoreFields = make([]string, 0)
			}
			patch.ignoreFields = append(patch.ignoreFields, strings.ToLower(oldElem.Type().Field(i).Name))
			continue
		}
	}

	patch.patchGen(old)

	return patch, nil
}
