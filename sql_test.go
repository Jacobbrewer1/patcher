package patcher

import (
	"errors"
	"fmt"
	"testing"

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
	s.Equal([]any{int64(1), "test"}, patch.args)
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
	s.Equal([]any{int64(1), nil}, patch.args)
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
	s.Equal([]any{int64(1), ""}, patch.args)
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
	s.Equal([]any{1, "1234 Main St"}, patch.args)
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

	s.Equal([]string{"Id = ?", "Name = ?"}, patch.fields)
	s.Equal([]any{int64(1), "test"}, patch.args)
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
	s.Equal([]any{int64(1), "test"}, patch.args)
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
		Id          int    `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
	}

	obj := testObj{
		Id:          73,
		Name:        "test",
		Description: "",
	}

	patch := NewSQLPatch(obj, WithIncludeZeroValues())

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

	patch := NewSQLPatch(obj, WithIncludeZeroValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(73), "test", ""}, patch.args)
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

	patch := NewSQLPatch(obj, WithIncludeZeroValues())

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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(73), "test", nil}, patch.args)
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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(0), "test", nil}, patch.args)
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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(0), nil, nil}, patch.args)
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

	patch := NewSQLPatch(obj, WithIncludeNilValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{int64(0), nil, "desc"}, patch.args)
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

	patch := NewSQLPatch(obj, WithIncludeNilValues(), WithIncludeZeroValues())

	s.Equal([]string{"id = ?", "name = ?", "description = ?"}, patch.fields)
	s.Equal([]any{nil, "", nil}, patch.args)
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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), nil, 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "", 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nAND name = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18, "john"}, args)

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
	mw2.On("WhereType").Return(WhereTypeOr)

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithWhere(mw2),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nOR name = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18, "john"}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nOR name = ?\nAND id = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18, "john", 1}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\nAND name = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18, "john"}, args)

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
	mj.On("Join").Return("JOIN table2 ON table1.id = table2.id", []any{})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
		WithJoin(mj),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
	s.NoError(err)
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\nSET id = ?, name = ?\nWHERE (1=1)\nAND (\nage = ?\n)", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(73), "test", "", 73}, args)
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
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(73), "test", "", 73}, args)
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
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.True(errors.Is(err, ErrNoFields))
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
		WithIncludeNilValues(),
		WithWhere(mw),
	)
	s.NoError(err)
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
		WithIncludeNilValues(),
		WithWhere(mw),
	)
	s.NoError(err)
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
		WithIncludeNilValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(73), "test", nil, 1}, args)
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
		WithIncludeNilValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(0), "test", nil, 1}, args)
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
		WithIncludeNilValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(0), nil, nil, 1}, args)
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
		WithIncludeNilValues(),
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.NoError(err)
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
		WithIncludeNilValues(),
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.NoError(err)
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
		WithIncludeNilValues(),
		WithIncludeZeroValues(),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("UPDATE test_table\nSET id = ?, name = ?, description = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{int64(73), "", nil, 1}, args)
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
	s.NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{int64(2), "test2"}, patch.args)
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
	s.NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{int64(2), nil}, patch.args)
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
	s.NoError(err)

	s.NotNil(patch)
	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
	s.Equal([]any{int64(2), ""}, patch.args)
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
	s.NoError(err)

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

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_ignoreNoChanges() {
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
	s.NoError(IgnoreNoChangesErr(err))
	s.Nil(patch)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_Success_ignoreNoChanges_wrapped() {
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
	if err != nil {
		err = IgnoreNoChangesErr(fmt.Errorf("wrapped: %w", err))
	}
	s.NoError(err)
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
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeZeroValues())
	s.True(errors.Is(err, ErrNoChanges))
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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

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

	patch, err := NewDiffSQLPatch(&obj, &obj2, WithTable("test_table"), WithWhere(mw), WithIncludeNilValues(), WithIncludeZeroValues())
	s.NoError(err)

	sqlStr, args, err := patch.GenerateSQL()
	s.NoError(err)

	s.Equal("UPDATE test_table\nSET description = ?, addr = ?\nWHERE (1=1)\nAND (\nid = ?\n)", sqlStr)
	s.Equal([]any{"", nil, 1}, args)
}
