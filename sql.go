package patcher

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
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
	resource = dereferenceIfPointer(resource)
	ensureStruct(resource)

	rType := reflect.TypeOf(resource)
	rVal := reflect.ValueOf(resource)
	n := rType.NumField()

	s.fields = make([]string, 0, n)
	s.args = make([]any, 0, n)

	for i := range n {
		fType := rType.Field(i)
		fVal := rVal.Field(i)
		tag := getTag(&fType, s.tagName)
		optsTag := fType.Tag.Get(TagOptsName)

		if s.shouldSkipField(&fType, fVal) {
			continue
		}

		var arg any = nil
		if fVal.Kind() == reflect.Ptr && fVal.IsNil() {
			if !s.shouldIncludeNil(optsTag) {
				continue
			}
		} else {
			arg = getValue(fVal)
		}

		s.fields = append(s.fields, tag+" = ?")
		s.args = append(s.args, arg)
	}
}

// GenerateSQL generates the SQL update statement and its arguments for the given resource.
// It creates a new SQLPatch instance with the provided options, processes the resource's fields,
// and constructs the SQL update statement along with the necessary arguments.
func GenerateSQL(resource any, opts ...PatchOpt) (sqlStr string, args []any, err error) {
	return NewSQLPatch(resource, opts...).GenerateSQL()
}

// GenerateSQL constructs the SQL update statement and its arguments.
// It validates the SQL generation process, builds the SQL update statement
// with the table name, join clauses, set clauses, and where clauses,
// and returns the final SQL string along with the arguments.
func (s *SQLPatch) GenerateSQL() (sqlStr string, args []any, err error) {
	if err := s.validateSQLGen(); err != nil {
		return "", nil, fmt.Errorf("validate SQL generation: %w", err)
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

	sqlArgs := s.joinArgs
	sqlArgs = append(sqlArgs, s.args...)
	sqlArgs = append(sqlArgs, s.whereArgs...)

	return sqlBuilder.String(), sqlArgs, nil
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
			patch.ignoreFields = append(patch.ignoreFields, oldElem.Type().Field(i).Name)
			continue
		}
	}

	patch.patchGen(old)

	return patch, nil
}
