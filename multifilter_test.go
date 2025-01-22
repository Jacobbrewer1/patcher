package patcher

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type multiFilterSuite struct {
	suite.Suite
}

func TestMultiFilterSuite(t *testing.T) {
	suite.Run(t, new(multiFilterSuite))
}

func (s *multiFilterSuite) TestNewMultiFilter_Add_Single() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("where", []any{"arg1", "arg2"})
	mf.Add(mw)

	sql, args := mf.Where()
	s.Equal("AND where\n", sql)
	s.Equal([]any{"arg1", "arg2"}, args)
}

func (s *multiFilterSuite) TestNewMultiFilter_Add_Multi() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("where", []any{"arg1", "arg2"})
	mf.Add(mw)

	mwTwo := NewMockWherer(s.T())
	mwTwo.On("Where").Return("whereTwo", []any{"arg3", "arg4"})
	mf.Add(mwTwo)

	sql, args := mf.Where()
	s.Equal("AND where\nAND whereTwo\n", sql)
	s.Equal([]any{"arg1", "arg2", "arg3", "arg4"}, args)
}

func (s *multiFilterSuite) TestNewMultiFilter_Add_WhereTyper() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("where", []any{"arg1", "arg2"})
	mf.Add(mw)

	mwt := NewMockWhereTyper(s.T())
	mwt.On("Where").Return("whereTwo", []any{"arg3", "arg4"})
	mwt.On("WhereType").Return(WhereTypeOr)
	mf.Add(mwt)

	sql, args := mf.Where()
	s.Equal("AND where\nOR whereTwo\n", sql)
	s.Equal([]any{"arg1", "arg2", "arg3", "arg4"}, args)
}
