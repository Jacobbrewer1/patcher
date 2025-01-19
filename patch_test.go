package patcher

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type newPatchSuite struct {
	suite.Suite
}

func TestNewPatchSuite(t *testing.T) {
	suite.Run(t, new(newPatchSuite))
}

func (s *newPatchSuite) TestNewPatch() {
	mpo := NewMockPatchOpt(s.T())
	mpo.On("Execute", mock.AnythingOfType("*patcher.SQLPatch"))

	_ = newPatchDefaults(mpo.Execute)

	mpo.AssertExpectations(s.T())
}
