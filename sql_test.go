package patcher

import (
	"testing"

	"github.com/Jacobbrewer1/patcher/utils"
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
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id_tag = ?", "name_tag = ?"}, patch.fields)
	s.Equal([]any{int64(1), "test"}, patch.args)
}

func (s *newSQLPatchSuite) TestNewSQLPatch_Success_noDbTag() {
	type testObj struct {
		Id   *int
		Name *string
	}

	obj := testObj{
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
	}

	patch := NewSQLPatch(obj)

	s.Equal([]string{"id = ?", "name = ?"}, patch.fields)
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
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
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
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
	}

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("age = ?", []any{18})

	sqlStr, args, err := GenerateSQL(obj,
		WithTable("test_table"),
		WithWhere(mw),
	)
	s.NoError(err)
	s.Equal("\n\t\tUPDATE test_table\n\t\t\n\t\tSET id = ?, name = ?\n\t\tWHERE 1\n\t\tAND age = ?\n\n\t", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_multipleWhere() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
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
	s.Equal("\n\t\tUPDATE test_table\n\t\t\n\t\tSET id = ?, name = ?\n\t\tWHERE 1\n\t\tAND age = ?\nAND name = ?\n\n\t", sqlStr)
	s.Equal([]any{int64(1), "test", 18, "john"}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_withJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
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
	s.Equal("\n\t\tUPDATE test_table\n\t\tJOIN table2 ON table1.id = table2.id\n\n\t\tSET id = ?, name = ?\n\t\tWHERE 1\n\t\tAND age = ?\n\n\t", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

	mw.AssertExpectations(s.T())
}

func (s *generateSQLSuite) TestGenerateSQL_Success_multipleJoin() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   utils.Ptr(1),
		Name: utils.Ptr("test"),
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
	s.Equal("\n\t\tUPDATE test_table\n\t\tJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\n\n\t\tSET id = ?, name = ?\n\t\tWHERE 1\n\t\tAND age = ?\n\n\t", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

	mw.AssertExpectations(s.T())
}
