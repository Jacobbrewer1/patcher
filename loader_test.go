package patcher

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type loadDiffSuite struct {
	suite.Suite

	l *loader
}

func TestLoadDiffSuite(t *testing.T) {
	suite.Run(t, new(loadDiffSuite))
}

func (s *loadDiffSuite) SetupTest() {
	s.l = newLoader()
}

func (s *loadDiffSuite) TearDownTest() {
	s.l = nil
}

func (s *loadDiffSuite) TestLoadDiff_Success() {
	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  26,
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(26, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_ZeroValue() {
	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  0,
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(25, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_NoNewValue() {
	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_OneNewField() {
	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Age: 26,
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(26, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_EmbeddedStruct() {
	type testStruct struct {
		Name    string
		Age     int
		Partner *testStruct
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Partner: &testStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
	s.Equal("Sarah", old.Partner.Name)
	s.Equal(24, old.Partner.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_EmbeddedStructWithNewValue() {
	type testStruct struct {
		Name    string
		Age     int
		Partner *testStruct
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Partner: &testStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	n.Partner.Name = "Sarah Brewer"
	n.Partner.Age = 25

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(25, old.Age)
	s.Equal("Sarah Brewer", old.Partner.Name)
	s.Equal(25, old.Partner.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_EmbeddedInheritedStruct() {
	type TestStruct struct {
		Name string
		Age  int
	}

	type testStruct2 struct {
		*TestStruct
		Partner *TestStruct
	}

	old := testStruct2{
		TestStruct: &TestStruct{
			Name: "John",
			Age:  25,
		},
	}

	n := testStruct2{
		TestStruct: &TestStruct{
			Name: "John Smith",
			Age:  26,
		},
		Partner: &TestStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(26, old.Age)
	s.Equal("Sarah", old.Partner.Name)
	s.Equal(24, old.Partner.Age)
}

func (s *loadDiffSuite) TestLoadDiff_FailureNotPointer() {
	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  26,
	}

	err := s.l.loadDiff(old, n)
	s.Error(err)
	s.Equal(ErrInvalidType, err)
}

// TestLoadDiff_Success_NilOldField ensures that a nil field in the old struct can be updated by the new struct.
func (s *loadDiffSuite) TestLoadDiff_Success_NilOldField() {
	type testStruct struct {
		Name    string
		Age     int
		Partner *testStruct
	}

	old := testStruct{
		Name:    "John",
		Age:     25,
		Partner: nil,
	}

	n := testStruct{
		Partner: &testStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
	s.Equal("Sarah", old.Partner.Name)
	s.Equal(24, old.Partner.Age)
}

// TestLoadDiff_Success_Slice ensures that slices are correctly copied over.
func (s *loadDiffSuite) TestLoadDiff_Success_Slice() {
	type testStruct struct {
		Tags []string
	}

	old := testStruct{
		Tags: []string{"tag1", "tag2"},
	}

	n := testStruct{
		Tags: []string{"tag3"},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal([]string{"tag3"}, old.Tags) // New slice overwrites old one
}

// TestLoadDiff_Success_DeeplyNestedStruct tests the handling of a deeply nested struct.
func (s *loadDiffSuite) TestLoadDiff_Success_DeeplyNestedStruct() {
	type InnerMost struct {
		Value string
	}
	type Inner struct {
		InnerMost InnerMost
	}
	type Outer struct {
		Inner Inner
	}

	old := Outer{
		Inner: Inner{
			InnerMost: InnerMost{
				Value: "Old Value",
			},
		},
	}

	n := Outer{
		Inner: Inner{
			InnerMost: InnerMost{
				Value: "New Value",
			},
		},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("New Value", old.Inner.InnerMost.Value)
}

// TestLoadDiff_Failure_UnexportedField tests if the function handles unexported fields correctly.
func (s *loadDiffSuite) TestLoadDiff_Failure_UnexportedField() {
	type testStruct struct {
		name string // unexported field, should not be set
		Age  int
	}

	old := testStruct{
		name: "OldName",
		Age:  25,
	}

	n := testStruct{
		name: "NewName",
		Age:  26,
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal(26, old.Age)
	s.Equal("OldName", old.name) // Name should remain unchanged because it's unexported
}

// TestLoadDiff_Failure_UnsupportedType ensures that types like channels return an error.
func (s *loadDiffSuite) TestLoadDiff_Failure_UnsupportedType() {
	type testStruct struct {
		Name    string
		Updates chan string // unsupported field type
	}

	old := testStruct{
		Name:    "John",
		Updates: make(chan string),
	}

	n := testStruct{
		Name: "John Smith",
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.NotNil(old.Updates) // Channel should not be nil as it started as a non-nil channel
}

func (s *loadDiffSuite) TestLoadDiff_Success_Include_Zeros() {
	l := s.l
	l.includeZeroValues = true

	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  0,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(0, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_Include_Zeros_false() {
	l := s.l
	l.includeZeroValues = false

	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  0,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(25, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_Include_Nil() {
	l := s.l
	l.includeNilValues = true

	type testStruct struct {
		Name    string
		Age     int
		Partner *testStruct
	}

	old := testStruct{
		Name: "John",
		Age:  25,
		Partner: &testStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	n := testStruct{
		Partner: nil,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
	s.Nil(old.Partner)
}

func (s *loadDiffSuite) TestLoadDiff_Success_Include_Nil_false() {
	l := s.l
	l.includeNilValues = false

	type testStruct struct {
		Name    string
		Age     int
		Partner *testStruct
	}

	old := testStruct{
		Name: "John",
		Age:  25,
		Partner: &testStruct{
			Name: "Sarah",
			Age:  24,
		},
	}

	n := testStruct{
		Partner: nil,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
	s.Equal("Sarah", old.Partner.Name)
	s.Equal(24, old.Partner.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_IgnoreFields() {
	l := s.l
	l.ignoreFields = []string{"name"}

	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  26,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(26, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_IgnoreFieldsFunc() {
	l := s.l
	l.ignoreFieldsFunc = func(fieldName string, oldValue, newValue any) bool {
		return fieldName == "name"
	}

	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  26,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(26, old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_IgnoreFieldsFuncAndIgnoreFields() {
	l := s.l
	l.ignoreFields = []string{"name"}
	l.ignoreFieldsFunc = func(fieldName string, oldValue, newValue any) bool {
		return fieldName == "name"
	}

	type testStruct struct {
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Name: "John Smith",
		Age:  26,
	}

	err := l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(26, old.Age)
}
