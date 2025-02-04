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

func (s *multiFilterSuite) TestNewMultiFilter_Add_Joiner() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("join", []any{"arg1", "arg2"})
	mf.Add(mj)

	sql, args := mf.Join()
	s.Equal("join\n", sql)
	s.Equal([]any{"arg1", "arg2"}, args)
}

func (s *multiFilterSuite) TestNewMultiFilter_Add_MultiJoiner() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("join", []any{"arg1", "arg2"})
	mf.Add(mj)

	mjTwo := NewMockJoiner(s.T())
	mjTwo.On("Join").Return("joinTwo", []any{"arg3", "arg4"})
	mf.Add(mjTwo)

	sql, args := mf.Join()
	s.Equal("join\njoinTwo\n", sql)
	s.Equal([]any{"arg1", "arg2", "arg3", "arg4"}, args)
}

func (s *multiFilterSuite) TestNewMultiFilter_Add_JoinerAndWherer() {
	mf := NewMultiFilter()
	s.NotNil(mf)

	mj := NewMockJoiner(s.T())
	mj.On("Join").Return("join", []any{"arg1", "arg2"})
	mf.Add(mj)

	mw := NewMockWherer(s.T())
	mw.On("Where").Return("where", []any{"arg3", "arg4"})
	mf.Add(mw)

	sql, args := mf.Join()
	s.Equal("join\n", sql)
	s.Equal([]any{"arg1", "arg2"}, args)

	sql, args = mf.Where()
	s.Equal("AND where\n", sql)
	s.Equal([]any{"arg3", "arg4"}, args)
}
