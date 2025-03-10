package patcher

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type newSQLPatchSuite struct {
	suite.Suite
}

func TestNewSQLPatchSuite(t *testing.T) {
	suite.Run(t, new(newSQLPatchSuite))
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_NoTableProvided() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
	s.Equal("test_obj", patch.table)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_Filter_Where() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockFilter(s.T())
	mf.On("Where").Return("test_where = ? and arg2_val = ?", []any{"arg1", "arg2"})

	patch := NewSQLPatch(obj, WithWhere(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)

	s.Equal("AND test_where = ? and arg2_val = ?\n", patch.whereSql.String())
	s.Equal([]any{"arg1", "arg2"}, patch.whereArgs)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_Filter_Join() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockFilter(s.T())
	mf.On("Join").Return("JOIN table2 ON table1.id = table2.id and arg2_val = ?", []any{"arg1"})

	patch := NewSQLPatch(obj, WithJoin(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)

	s.Equal("JOIN table2 ON table1.id = table2.id and arg2_val = ?\n", patch.joinSql.String())
	s.Equal([]any{"arg1"}, patch.joinArgs)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_Filter_JoinerAndWhere() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockFilter(s.T())
	mf.On("Join").Return("JOIN table2 ON table1.id = table2.id and arg2_val = ?", []any{"arg1"})
	mf.On("Where").Return("where", []any{"arg1", "arg2"})

	patch := NewSQLPatch(obj, WithFilter(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)

	s.Equal("JOIN table2 ON table1.id = table2.id and arg2_val = ?\n", patch.joinSql.String())
	s.Equal([]any{"arg1"}, patch.joinArgs)

	s.Equal("AND where\n", patch.whereSql.String())
	s.Equal([]any{"arg1", "arg2"}, patch.whereArgs)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_MultiFilter() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockMultiFilter(s.T())
	mf.On("Where").Return("where", []any{"arg1", "arg2"})

	patch := NewSQLPatch(obj, WithWhere(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_MultiFilter_Joiner() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockMultiFilter(s.T())
	mf.On("Join").Return("JOIN table2 ON table1.id = table2.id", nil)

	patch := NewSQLPatch(obj, WithJoin(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_MultiFilter_JoinerAndWhere() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mf := NewMockMultiFilter(s.T())
	mf.On("Join").Return("JOIN table2 ON table1.id = table2.id", nil)
	mf.On("Where").Return("where", []any{"arg1", "arg2"})

	patch := NewSQLPatch(obj, WithJoin(mf), WithWhere(mf))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_WhereString() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj, WithWhereStr("age = ?", 18))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)

	s.Equal("AND age = ?\n", patch.whereSql.String())
	s.Equal([]any{18}, patch.whereArgs)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_JoinString() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj, WithJoinStr("JOIN table2 ON table1.id = table2.id"))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)

	s.Equal("JOIN table2 ON table1.id = table2.id\n", patch.joinSql.String())
	s.Empty(patch.joinArgs)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Fields_Args_Getters() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.Fields())
	s.Equal([]any{1, "test"}, patch.Args())
}

func (s *newSQLPatchSuite) TestPatchGen_AllTypes() {
	type testObj struct {
		IntVal        int
		Int8Val       int8
		Int16Val      int16
		Int32Val      int32
		Int64Val      int64
		UintVal       uint
		Uint8Val      uint8
		Uint16Val     uint16
		Uint32Val     uint32
		Uint64Val     uint64
		UintptrVal    uintptr
		Float32Val    float32
		Float64Val    float64
		Complex64Val  complex64
		Complex128Val complex128
		StringVal     string
		BoolVal       bool
	}

	obj := testObj{
		IntVal:        1,
		Int8Val:       2,
		Int16Val:      3,
		Int32Val:      4,
		Int64Val:      5,
		UintVal:       6,
		Uint8Val:      7,
		Uint16Val:     8,
		Uint32Val:     9,
		Uint64Val:     10,
		UintptrVal:    11,
		Float32Val:    12.34,
		Float64Val:    56.78,
		Complex64Val:  complex(1, 2),
		Complex128Val: complex(3, 4),
		StringVal:     "test",
		BoolVal:       true,
	}

	patch := NewSQLPatch(obj)

	expectedFields := []string{
		"intval = ?", "int8val = ?", "int16val = ?", "int32val = ?", "int64val = ?",
		"uintval = ?", "uint8val = ?", "uint16val = ?", "uint32val = ?", "uint64val = ?", "uintptrval = ?",
		"float32val = ?", "float64val = ?", "stringval = ?", "boolval = ?",
	}
	expectedArgs := []any{
		1, int8(2), int16(3), int32(4), int64(5),
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10), uintptr(11),
		float32(12.34), 56.78, "test", true,
	}

	s.Equal(expectedFields, patch.fields)
	s.Equal(expectedArgs, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_MultipleTags() {
	type testObj struct {
		Id   *int    `db:"id_tag,pk"`
		Name *string `db:"name_tag,unique"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_DifferentTag() {
	type testObj struct {
		Id   *int    `tagged:"id_tag,pk"`
		Name *string `tagged:"name_tag,unique"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj, WithTagName("tagged"))

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_Struct_opt_IncludeNilFields() {
	type testObj struct {
		Id   *int    `db:"id_tag"`
		Name *string `db:"name_tag" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: nil,
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_Struct_opt_IncludeZeroFields() {
	type testObj struct {
		Id   int    `db:"id_tag"`
		Name string `db:"name_tag" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   1,
		Name: "",
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{1, ""}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Skip() {
	type testObj struct {
		Id      *int   `db:"id_tag" patcher:"-"`
		Name    string `db:"name_tag" patcher:"-"`
		Deleted bool   `db:"deleted"`
		Address string `db:"address"`
	}

	obj := testObj{
		Id:      ptr(1),
		Name:    "test",
		Deleted: true,
		Address: "1234 Main St",
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"deleted = ?", "address = ?"}, patch.fields)
	s.Equal([]any{true, "1234 Main St"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_noDbTag() {
	type testObj struct {
		Id   *int
		Name *string
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_noPointer() {
	type testObj struct {
		Id   int    `db:"id"`
		Name string `db:"name"`
	}

	obj := testObj{
		Name: "test",
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"name = ?"}, patch.fields)
	s.Equal([]any{"test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_PointedObj() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := &testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{1, "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_fail_notStruct() {
	obj := 1

	s.Panics(func() {
		_ = NewSQLPatch(obj)
	})
}

func (s *newSQLPatchSuite) TestNewSQLPatch_fail_noFields() {
	type testObj struct{}

	obj := testObj{}

	// This will return a patch object with no fields
	patch := NewSQLPatch(obj)

	s.Equal([]string{}, patch.fields)
	s.Equal([]any{}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeZeroValue() {
	type testObj struct {
		Id          int64  `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
	}

	obj := testObj{
		Id:          int64(73),
		Name:        "test",
		Description: "",
	}

	patch := NewSQLPatch(obj, WithIncludeZeroValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(73), "test", ""}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeZeroValue_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        ptr("test"),
		Description: ptr(""),
	}

	patch := NewSQLPatch(obj, WithIncludeZeroValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{73, "test", ""}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeZeroValue_PointerNil() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeZeroValues(true))

	// Nothing should be included as we are including zero values and all fields are nil
	s.Empty(patch.fields)
	s.Empty(patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{nil, nil, nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{nil, nil, nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_PointerWithValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        ptr("test"),
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{73, "test", nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_PointerWithZeroValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        ptr("test"),
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{0, "test", nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_PointerWithZeroValueAndNil() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        nil,
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{0, nil, nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_PointerWithZeroValueAndNilAndValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        nil,
		Description: ptr("desc"),
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{0, nil, "desc"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IncludeNilValue_IncludeZeroValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true), WithIncludeZeroValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{nil, "", nil}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_WithDB() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	// Setup mock database
	db := &sql.DB{}

	patch := NewSQLPatch(obj, WithDB(db), WithIncludeNilValues(true), WithIncludeZeroValues(true))

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{nil, "", nil}, patch.args)

	s.Equal(db, patch.db)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IgnoredFields() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	patch := NewSQLPatch(obj, WithIncludeNilValues(true), WithIncludeZeroValues(true), WithIgnoredFields("Id", "Description"))

	s.Equal([]string{"name = ?"}, patch.fields)
	s.Equal([]any{""}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Failure_FuncArg() {
	type testObj struct {
		Id       *int   `db:"id"`
		Name     string `db:"name"`
		Runnable func() `db:"func"`
	}

	obj := testObj{
		Id:   ptr(73),
		Name: "Test Name",
		Runnable: func() {
			fmt.Println("Hello")
		},
	}

	p := new(SQLPatch)
	s.NotPanics(func() {
		p = NewSQLPatch(obj)
	})
	s.NotNil(p)

	s.Equal([]string{"id = ?", "name = ?"}, p.fields)
	s.Equal([]any{73, "Test Name"}, p.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_IgnoredFieldsFunc() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	ignoreFunc := NewMockIgnoreFieldsFunc(s.T())
	ignoreFunc.On("Execute", mock.AnythingOfType("*reflect.StructField")).Return(func(field *reflect.StructField) bool {
		return field.Name == "Id" || field.Name == "Description"
	})

	patch := NewSQLPatch(obj, WithIncludeNilValues(true), WithIncludeZeroValues(true), WithIgnoredFieldsFunc(ignoreFunc.Execute))

	s.Equal([]string{"name = ?"}, patch.fields)
	s.Equal([]any{""}, patch.args)
}

type generateSQLSuite struct {
	suite.Suite
}

func TestGenerateSQLSuite(t *testing.T) {
	suite.Run(t, new(generateSQLSuite))
}

func (s *generateSQLSuite) TestGenerateSQL_Success() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_NoTableProvided() {
	type testTable struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testTable{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

type testFilter struct{}

func (f *testFilter) Where() (sqlStr string, args []any) {
	return "age = ?", []any{18}
}

func (f *testFilter) Join() (sqlStr string, args []any) {
	return "JOIN table2 ON table1.id = table2.id", nil
}

func (s *generateSQLSuite) TestGenerateSQL_Success_WhereAndJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithFilter(new(testFilter)),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_WhereString() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhereStr("age = ?", 18),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_JoinString() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithJoinStr("JOIN table2 ON table1.id = table2.id"),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_NoWhereArgs() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age > 18", nil)

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage > 18\n)", sqlStr)
	s.Equal([]any{1, "test"}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_Stuct_opt_IncludeNilFields() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, nil, 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_Struct_opt_IncludeZeroFields() {
	type testObj struct {
		Id   int    `db:"id"`
		Name string `db:"name" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   1,
		Name: "",
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_multipleWhere() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mw2 := NewMockWherer(s.T())
	mw2.On("Where").Return("name = ?", []any{"john"})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithWhere(mw2),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nAND name = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18, "john"}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_orWhere() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mw2 := NewMockWhereTyper(s.T())
	mw2.On("Where").Return("name = ?", []any{"john"})
	mw2.On("WhereType").Return(func() WhereType {
		return WhereTypeOr
	})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithWhere(mw2),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nOR name = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18, "john"}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_andOrWhere() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mw2 := NewMockWhereTyper(s.T())
	mw2.On("Where").Return("name = ?", []any{"john"})
	mw2.On("WhereType").Return(WhereTypeOr)

	mw3 := NewMockWherer(s.T())
	mw3.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithWhere(mw2),
		WithWhere(mw3),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nOR name = ?\nAND id = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18, "john", 1}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_invalidWhereType() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mw2 := NewMockWhereTyper(s.T())
	mw2.On("Where").Return("name = ?", []any{"john"})
	mw2.On("WhereType").Return(WhereType("invalid"))

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithWhere(mw2),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nAND name = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18, "john"}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_withJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("JOIN table2 ON table1.id = table2.id", nil)

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithJoin(mj),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_multipleJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("JOIN table2 ON table1.id = table2.id", []any{})

	mj2 := NewMockJoiner(s.T())
	mj2.On("Join").Return("JOIN table3 ON table1.id = table3.id", []any{})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithJoin(mj),
		WithJoin(mj2),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_withJoinAndWhere() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("JOIN table2 ON table1.id = table2.id", []any{})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithJoin(mj),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_withJoinAndWhereAndJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("JOIN table2 ON table1.id = table2.id", []any{})

	mj2 := NewMockJoiner(s.T())
	mj2.On("Join").Return("JOIN table3 ON table1.id = table3.id", []any{})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithJoin(mj),
		WithJoin(mj2),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{1, "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesZeroValues() {
	type testObj struct {
		Id          int    `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
	}

	obj := testObj{
		Id:          73,
		Name:        "test",
		Description: "",
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{73})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{73, "test", "", 73}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesZeroValues_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        ptr("test"),
		Description: ptr(""),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{73})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{73, "test", "", 73}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesZeroValues_PointerNil() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().ErrorIs(err, ErrNoFields)
	s.Empty(sqlStr)
	s.Empty(args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, nil, nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, nil, nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_PointerWithValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        ptr("test"),
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{73, "test", nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_PointerWithZeroValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        ptr("test"),
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{0, "test", nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_PointerWithZeroValueAndNil() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        nil,
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{0, nil, nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_IncludesZeroValues() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, "", nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_IncludesZeroValues_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, "", nil, 1}, args)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_IncludesNilValues_IncludesZeroValues_PointerWithValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        "",
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithIncludeNilValues(true),
		WithIncludeZeroValues(true),
		WithWhere(mw),
	)
	s.Require().NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{73, "", nil, 1}, args)
}

type NewDiffSQLPatchSuite struct {
	suite.Suite
}

func TestNewDiffSQLPatchSuite(t *testing.T) {
	suite.Run(t, new(NewDiffSQLPatchSuite))
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	obj2 := testObj{
		Id:   ptr(2),
		Name: ptr("test2"),
	}

	patch, err := NewDiffSQLPatch(&obj, &obj2)
	s.Require().NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{2, "test2"}, patch.args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_StructOpt_IncludeNilFields() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	obj2 := testObj{
		Id:   ptr(2),
		Name: nil,
	}

	patch, err := NewDiffSQLPatch(&obj, &obj2)
	s.Require().NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{2, nil}, patch.args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_StructOpt_IncludeZeroFields() {
	type testObj struct {
		Id   int    `db:"id"`
		Name string `db:"name" patcher:"omitempty"`
	}

	obj := testObj{
		Id:   1,
		Name: "test",
	}

	obj2 := testObj{
		Id:   2,
		Name: "",
	}

	patch, err := NewDiffSQLPatch(&obj, &obj2)
	s.Require().NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{2, ""}, patch.args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_singleFieldUpdated() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
		Desc string  `db:"desc"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
		Desc: "desc",
	}

	obj2 := testObj{
		Id:   ptr(1),
		Name: ptr("test2"),
		Desc: "desc",
	}

	patch, err := NewDiffSQLPatch(&obj, &obj2)
	s.Require().NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"name = ?"}, patch.fields)
	s.Equal([]any{"test2"}, patch.args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_noChange() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	obj2 := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	patch, err := NewDiffSQLPatch(&obj, &obj2)
	s.Equal(ErrNoChanges, err)
	s.Nil(patch)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_fail_notStruct() {
	obj := 1

	_, err := NewDiffSQLPatch(&obj, &obj)
	s.Error(err)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	obj2 := testObj{
		Id:   ptr(1),
		Name: ptr("test2"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{"test2", 18}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_NoTableProvided() {
	type testTable struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testTable{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	obj2 := testTable{
		Id:   ptr(1),
		Name: ptr("test2"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithWhere(mw))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{"test2", 18}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_ValueField() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
		Desc string  `db:"desc"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
		Desc: "desc",
	}

	obj2 := testObj{
		Id:   ptr(1),
		Name: ptr("test2"),
		Desc: "desc",
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{"test2", 18}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_ValueFieldUpdated() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
		Desc string  `db:"desc"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
		Desc: "desc",
	}

	obj2 := testObj{
		Id:   ptr(1),
		Name: ptr("test2"),
		Desc: "desc2",
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?, desc = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{"test2", "desc2", 18}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeZeroValues() {
	type testObj struct {
		Id          int    `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
	}

	obj := testObj{
		Id:          73,
		Name:        "test",
		Description: "desc",
	}

	obj2 := testObj{
		Id:          73,
		Name:        "test2",
		Description: "",
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{73})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{"test2", "", 73}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeZeroValues_Pointer() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
		Addr        *string `db:"addr"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        ptr("test"),
		Description: ptr("desc"),
		Addr:        ptr("addr"),
	}

	obj2 := testObj{
		Id:          ptr(73),
		Name:        ptr("test2"),
		Description: ptr(""),
		Addr:        nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{73})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{"test2", "", 73}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeZeroValues_PointerNil() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        *string `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	obj2 := testObj{
		Id:          nil,
		Name:        nil,
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues(true))
	s.Require().ErrorIs(err, ErrNoChanges)
	s.Nil(patch)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeNilValues() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(11),
		Name:        "test",
		Description: ptr("desc"),
	}

	obj2 := testObj{
		Id:          nil,
		Name:        "",
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET id = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, nil, 1}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeNilValues_PointerWithValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        "",
		Description: ptr("desc"),
	}

	obj2 := testObj{
		Id:          ptr(73),
		Name:        "",
		Description: nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{nil, 1}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeNilValues_PointerWithZeroValue() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description *string `db:"description"`
	}

	obj := testObj{
		Id:          ptr(0),
		Name:        "",
		Description: ptr("desc"),
	}

	obj2 := testObj{
		Id:          ptr(0),
		Name:        "",
		Description: ptr(""),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{"", 1}, args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_SqlGen_IncludeNilValues_IncludesZeroValues() {
	type testObj struct {
		Id          *int    `db:"id"`
		Name        string  `db:"name"`
		Description string  `db:"description"`
		Addr        *string `db:"addr"`
	}

	obj := testObj{
		Id:          ptr(73),
		Name:        "John",
		Description: "desc",
		Addr:        ptr(""),
	}

	obj2 := testObj{
		Id:          ptr(73),
		Name:        "John",
		Description: "",
		Addr:        nil,
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("id = ?", []any{1})

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues(true), WithIncludeZeroValues(true))
	s.Require().NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.Require().NoError(err)

	s.Equal("UPDATE test_table\nSET description = ?, addr = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{"", nil, 1}, args)
}
