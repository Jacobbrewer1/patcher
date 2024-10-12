package inserter

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

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

	b := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources))

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

	b := NewBatch(WithTable("temp"), WithResources(resources))

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

	b := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources))

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

	b := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources))

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

	b := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}

func (s *newBatchSuite) TestNewBatch_noResources() {
	b := NewBatch(WithTable("temp"), WithTagName("db"))

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

	b := NewBatch(WithTagName("db"), WithResources(resources))

	s.Require().Len(b.Fields(), 2)
	s.Require().Len(b.Args(), 10)
}

func (s *newBatchSuite) TestNewBatch_noTable_noResources() {
	b := NewBatch(WithTagName("db"))

	s.Require().Len(b.Fields(), 0)
	s.Require().Len(b.Args(), 0)
}
