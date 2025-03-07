package patcher

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPointerToStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"Pointer to struct", &struct{}{}, true},
		{"Nil pointer to struct", (*struct{})(nil), false},
		{"Non-pointer value", struct{}{}, false},
		{"Pointer to non-struct", new(int), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := isPointerToStruct(tt.value)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestDereferenceIfPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resource any
		expected any
	}{
		{"Non-pointer value", 42, 42},
		{"Pointer to value", ptr(42), 42},
		{"Nil pointer", (*int)(nil), (*int)(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := dereferenceIfPointer(tt.resource)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestEnsureStruct(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		resource    any
		shouldPanic bool
	}{
		{"Struct", struct{}{}, false},
		{"Pointer to Struct", &struct{}{}, true},
		{"Non-struct", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.shouldPanic {
				require.Panics(t, func() { ensureStruct(tt.resource) })
			} else {
				require.NotPanics(t, func() { ensureStruct(tt.resource) })
			}
		})
	}
}

func TestGetTag(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Field1 string `custom:"field1_tag"`
		Field2 string
	}

	tests := []struct {
		name     string
		field    string
		tagName  string
		expected string
	}{
		{"Field with tag", "Field1", "custom", "field1_tag"},
		{"Field without tag", "Field2", "custom", "field2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, _ := reflect.TypeOf(TestStruct{}).FieldByName(tt.field)
			actual := getTag(&f, tt.tagName)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected any
	}{
		{"Non-pointer value", 42, 42},
		{"Pointer to value", ptr(42), 42},
		{"Nil pointer", (*int)(nil), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := reflect.ValueOf(tt.value)
			actual := getValue(val)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestIsValidType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"Bool", true, true},
		{"Int", 1, true},
		{"Int8", int8(1), true},
		{"Int16", int16(1), true},
		{"Int32", int32(1), true},
		{"Int64", int64(1), true},
		{"Uint", uint(1), true},
		{"Uint8", uint8(1), true},
		{"Uint16", uint16(1), true},
		{"Uint32", uint32(1), true},
		{"Uint64", uint64(1), true},
		{"Uintptr", uintptr(1), true},
		{"Float32", float32(1.0), true},
		{"Float64", 1.0, true},
		{"String", "test", true},
		{"Struct", struct{}{}, true},
		{"Pointer", new(int), true},
		{"InvalidType", complex(1, 1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := reflect.ValueOf(tt.value)
			actual := IsValidType(val)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetTableName(t *testing.T) {
	t.Parallel()

	type TestStruct struct{}
	type TestStructWithPtr struct{}

	tests := []struct {
		name     string
		resource any
		expected string
	}{
		{
			name:     "Struct",
			resource: TestStruct{},
			expected: "test_struct",
		},
		{
			name:     "Pointer to Struct",
			resource: &TestStructWithPtr{},
			expected: "test_struct_with_ptr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := getTableName(tt.resource)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "TestToSnakeCase",
			expected: "test_to_snake_case",
		},
		{
			name:     "TestToSnakeCaseWithNumbers123",
			expected: "test_to_snake_case_with_numbers123",
		},
		{
			name:     "TestToSnakeCaseWithNumbers123AndUppercase",
			expected: "test_to_snake_case_with_numbers123_and_uppercase",
		},
		{
			name:     "TestToSnakeCaseWithNumbers123AndUppercaseAndSymbols!@#",
			expected: "test_to_snake_case_with_numbers123_and_uppercase_and_symbols!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := toSnakeCase(tt.name)
			require.Equal(t, tt.expected, actual)
		})
	}
}
