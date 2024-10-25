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

func (s *loadDiffSuite) TestLoadDiff_Success_Pointed_Fields() {
	s.l.includeNilValues = true

	type testStruct struct {
		Name *string
		Age  *int
	}

	old := testStruct{
		Name: ptr("John"),
		Age:  ptr(25),
	}

	n := testStruct{
		Name: ptr("John Smith"),
		Age:  ptr(26),
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", *old.Name)
	s.Equal(26, *old.Age)
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

func (s *loadDiffSuite) TestLoadDiff_Success_EmbeddedStruct_Reverse() {
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
		Partner: &testStruct{
			Name: "Sarah Thompson",
			Age:  27,
		},
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John", old.Name)
	s.Equal(25, old.Age)
	s.Equal("Sarah Thompson", old.Partner.Name)
	s.Equal(27, old.Partner.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_EmbeddedStruct_NotPointed() {
	type testEmbed struct {
		Name string
		Age  int
	}

	type testStruct struct {
		Name    string
		Age     int
		Partner testEmbed
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		Partner: testEmbed{
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

func (s *loadDiffSuite) TestLoadDiff_Success_InheritedStruct_NotPointed() {
	type TestEmbed struct {
		Description string
	}

	type testStruct struct {
		TestEmbed
		Name string
		Age  int
	}

	old := testStruct{
		Name: "John",
		Age:  25,
	}

	n := testStruct{
		TestEmbed: TestEmbed{
			Description: "Some description",
		},
		Name: "John Smith",
		Age:  26,
	}

	err := s.l.loadDiff(&old, &n)
	s.NoError(err)
	s.Equal("John Smith", old.Name)
	s.Equal(26, old.Age)
	s.Equal("Some description", old.Description)
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

func (s *loadDiffSuite) TestLoadDiff_Success_Blank_Except_Id() {
	l := s.l
	l.includeZeroValues = true
	l.ignoreFields = []string{"id"}

	type testStruct struct {
		ID   int
		Name string
		Age  *int
	}

	old := &testStruct{
		ID:   17345,
		Name: "some text",
		Age:  ptr(25),
	}

	n := &testStruct{
		ID:   0,
		Name: "",
		Age:  nil,
	}

	err := l.loadDiff(old, n)
	s.NoError(err)
	s.Equal(17345, old.ID)
	s.Equal("", old.Name)
	s.Equal(25, *old.Age)
}

func (s *loadDiffSuite) TestLoadDiff_Success_Skip_Priority_Check() {
	l := s.l
	l.includeZeroValues = true
	l.ignoreFields = []string{"id"}

	type testStruct struct {
		ID          int
		Name        string `patcher:"-"`
		Age         *int
		BankBalance int
	}

	old := &testStruct{
		ID:          17345,
		Name:        "some text",
		Age:         ptr(25),
		BankBalance: 1000,
	}

	n := &testStruct{
		ID:          0,
		Name:        "John Smith",
		Age:         nil,
		BankBalance: 0,
	}

	err := l.loadDiff(old, n)
	s.NoError(err)
	s.Equal(17345, old.ID)
	s.Equal("some text", old.Name)
	s.Nil(old.Age)
	s.Equal(0, old.BankBalance)
}

func (s *loadDiffSuite) TestLoadDiff_DefaultBehaviour() {
	type testStruct struct {
		ID   int
		Name string
		Age  *int
		Addr string
	}

	old := &testStruct{
		ID:   17345,
		Name: "some text",
		Age:  ptr(25),
		Addr: "",
	}

	n := &testStruct{
		ID:   0,
		Name: "John Smith",
		Age:  nil,
		Addr: "some address",
	}

	err := LoadDiff(old, n)
	s.NoError(err)
	s.Equal(17345, old.ID)
	s.Equal("John Smith", old.Name)
	s.Equal(25, *old.Age)
	s.Equal("some address", old.Addr)
}

func (s *loadDiffSuite) TestLoadDiff_IgnoreTags() {
	type testStruct struct {
		ID    int    `patcher:"-"`
		Name  string `patcher:"-"`
		Email string
		Age   *int
		Addr  string
	}

	old := &testStruct{
		ID:    17345,
		Name:  "some text",
		Email: "some email",
		Age:   ptr(25),
		Addr:  "",
	}

	n := &testStruct{
		ID:    0,
		Name:  "John Smith",
		Email: "some other email",
		Age:   nil,
		Addr:  "some address",
	}

	err := LoadDiff(old, n)
	s.NoError(err)
	s.Equal(17345, old.ID)
	s.Equal("some text", old.Name)
	s.Equal(25, *old.Age)
	s.Equal("some address", old.Addr)
	s.Equal("some other email", old.Email)
}
