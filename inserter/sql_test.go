package inserter

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func ptr[T any](v T) *T {
	return &v
}

type newBatchSuite struct {
	suite.Suite
}

func TestNewBatch(t *testing.T) {
	suite.Run(t, new(newBatchSuite))
}

func (s *newBatchSuite) TestNewBatch_Success() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	b := NewBatch(resources, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_Success_WithPointedFields() {
	type temp struct {
		ID         *int    `db:"id"`
		Name       *string `db:"name"`
		unexported string  `db:"unexported"`
	}

	resources := []any{
		&temp{ID: ptr(1), Name: ptr("test")},
		&temp{ID: ptr(2), Name: ptr("test2")},
		&temp{ID: ptr(3), Name: ptr("test3")},
		&temp{ID: ptr(4), Name: ptr("test4")},
		&temp{ID: ptr(5), Name: ptr("test5"), unexported: "test"},
	}

	b := NewBatch(resources, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_noDbTag() {
	type temp struct {
		ID         int
		Name       string
		unexported string
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	b := NewBatch(resources, WithTable("temp"))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_notPointer() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		temp{ID: 1, Name: "test"},
		temp{ID: 2, Name: "test2"},
		temp{ID: 3, Name: "test3"},
		temp{ID: 4, Name: "test4"},
		temp{ID: 5, Name: "test5", unexported: "test"},
	}

	b := NewBatch(resources, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_notStruct() {
	resources := []any{
		"test",
		"test2",
		"test3",
		"test4",
		"test5",
	}

	b := NewBatch(resources, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}

func (s *newBatchSuite) TestNewBatch_noFields() {
	type temp struct {
		unexported string
	}

	resources := []any{
		&temp{unexported: "test"},
		&temp{unexported: "test2"},
		&temp{unexported: "test3"},
		&temp{unexported: "test4"},
		&temp{unexported: "test5"},
	}

	b := NewBatch(resources, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}

func (s *newBatchSuite) TestNewBatch_noResources() {
	b := NewBatch(nil, WithTable("temp"), WithTagName("db"))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}

func (s *newBatchSuite) TestNewBatch_noTable() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	b := NewBatch(resources, WithTagName("db"))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_noTable_noResources() {
	b := NewBatch(nil, WithTagName("db"))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}

type generateSQLSuite struct {
	suite.Suite
}

func TestGenerateSQLSuite(t *testing.T) {
	suite.Run(t, new(generateSQLSuite))
}

func (s *generateSQLSuite) TestGenerateSQL_Success() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	sql, args, err := NewBatch(resources, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().NoError(err)

	s.Require().Equal("INSERT INTO temp (id, name) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?)", sql)
	s.Require().Len(args, 10)
}

func (s *generateSQLSuite) TestGenerateSQL_noDbTag() {
	type temp struct {
		ID         int
		Name       string
		unexported string
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	sql, args, err := NewBatch(resources, WithTable("temp")).GenerateSQL()
	s.Require().NoError(err)

	s.Require().Equal("INSERT INTO temp (id, name) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?)", sql)
	s.Require().Len(args, 10)
}

func (s *generateSQLSuite) TestGenerateSQL_notPointer() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		temp{ID: 1, Name: "test"},
		temp{ID: 2, Name: "test2"},
		temp{ID: 3, Name: "test3"},
		temp{ID: 4, Name: "test4"},
		temp{ID: 5, Name: "test5", unexported: "test"},
	}

	sql, args, err := NewBatch(resources, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().NoError(err)

	s.Require().Equal("INSERT INTO temp (id, name) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?)", sql)
	s.Require().Len(args, 10)
}

func (s *generateSQLSuite) TestGenerateSQL_notStruct() {
	resources := []any{
		"test",
		"test2",
		"test3",
		"test4",
		"test5",
	}

	sql, args, err := NewBatch(resources, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().Equal(ErrNoFields, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_noFields() {
	type temp struct {
		unexported string
	}

	resources := []any{
		&temp{unexported: "test"},
		&temp{unexported: "test2"},
		&temp{unexported: "test3"},
		&temp{unexported: "test4"},
		&temp{unexported: "test5"},
	}

	sql, args, err := NewBatch(resources, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().Equal(ErrNoFields, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_noResources() {
	sql, args, err := NewBatch(nil, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().Equal(ErrNoFields, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_noTable() {
	type temp struct {
		ID         int    `db:"id"`
		Name       string `db:"name"`
		unexported string `db:"unexported"`
	}

	resources := []any{
		&temp{ID: 1, Name: "test"},
		&temp{ID: 2, Name: "test2"},
		&temp{ID: 3, Name: "test3"},
		&temp{ID: 4, Name: "test4"},
		&temp{ID: 5, Name: "test5", unexported: "test"},
	}

	sql, args, err := NewBatch(resources, WithTagName("db")).GenerateSQL()
	s.Require().Equal(ErrNoTable, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_noTable_noResources() {
	sql, args, err := NewBatch(nil, WithTagName("db"), WithTable("temp")).GenerateSQL()
	s.Require().Equal(ErrNoFields, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_noTable_noFields() {
	type temp struct {
		unexported string
	}

	resources := []any{
		&temp{unexported: "test"},
		&temp{unexported: "test2"},
		&temp{unexported: "test3"},
		&temp{unexported: "test4"},
		&temp{unexported: "test5"},
	}

	sql, args, err := NewBatch(resources, WithTagName("db"), WithTable("temp")).GenerateSQL()
	s.Require().Equal(ErrNoFields, err)

	s.Require().Equal("", sql)
	s.Require().Len(args, 0)
}

func (s *generateSQLSuite) TestGenerateSQL_Success_WithPointedFields() {
	type temp struct {
		ID         *int    `db:"id"`
		Name       *string `db:"name"`
		unexported string  `db:"unexported"`
	}

	resources := []any{
		&temp{ID: ptr(1), Name: ptr("test")},
		&temp{ID: nil, Name: ptr("test2")},
		&temp{ID: ptr(3), Name: ptr("test3")},
		&temp{ID: ptr(4), Name: ptr("test4")},
		&temp{ID: ptr(5), Name: ptr("test5"), unexported: "test"},
	}

	sql, args, err := NewBatch(resources, WithTable("temp"), WithTagName("db")).GenerateSQL()
	s.Require().NoError(err)

	s.Require().Equal("INSERT INTO temp (id, name) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?)", sql)
	s.Require().Len(args, 10)

	expectedArgs := []any{resources[0].(*temp).ID, resources[0].(*temp).Name, (*int)(nil), resources[1].(*temp).Name, resources[2].(*temp).ID, resources[2].(*temp).Name, resources[3].(*temp).ID, resources[3].(*temp).Name, resources[4].(*temp).ID, resources[4].(*temp).Name}
	s.Require().Equal(expectedArgs, args)
}
