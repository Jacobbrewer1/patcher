package patcher

import (
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
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE 1\nAND age = ?\n", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

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
	s.Equal("UPDATE test_table\nSET id = ?, name = ?\nWHERE 1\nAND age = ?\nAND name = ?\n", sqlStr)
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
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nSET id = ?, name = ?\nWHERE 1\nAND age = ?\n", sqlStr)
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
	s.Equal("UPDATE test_table\nJOIN table2 ON table1.id = table2.id\nJOIN table3 ON table1.id = table3.id\nSET id = ?, name = ?\nWHERE 1\nAND age = ?\n", sqlStr)
	s.Equal([]any{int64(1), "test", 18}, args)

	mw.AssertExpectations(s.T())
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
	s.NoError(err)

	s.NotNil(patch)
	s.Equal([]string{}, patch.fields)
	s.Equal([]any{}, patch.args)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_fail_notStruct() {
	obj := 1

	_, err := NewDiffSQLPatch(obj, obj)
	s.Error(err)
}

func (s *NewDiffSQLPatchSuite) TestNewDiffSQLPatch_fail_notPointer() {
	type testObj struct {
		Id   *int    `db:"id"`
		Name *string `db:"name"`
	}

	obj := testObj{
		Id:   ptr(1),
		Name: ptr("test"),
	}

	_, err := NewDiffSQLPatch(obj, obj)
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

	s.Equal("UPDATE test_table\nSET name = ?\nWHERE 1\nAND age = ?\n", sqlStr)
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

	s.Equal("UPDATE test_table\nSET name = ?\nWHERE 1\nAND age = ?\n", sqlStr)
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

	s.Equal("UPDATE test_table\nSET name = ?, desc = ?\nWHERE 1\nAND age = ?\n", sqlStr)
	s.Equal([]any{"test2", "desc2", 18}, args)
}
