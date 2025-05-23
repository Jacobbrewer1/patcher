package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jacobbrewer1/patcher"
)

type Something struct {
	Number             int
	Text               string
	PrePopulated       string
	NewText            string
	ZeroValue          int
	NilValue           *string
	IgnoredField       string
	IgnoredFieldTwo    string
	IgnoredFieldByFunc string
	IgnoredByTag       string `patcher:"-"`
}

func main() {
	str := "nil value"

	s := Something{
		Number:             5,
		Text:               "Hello",
		PrePopulated:       "PrePopulated",
		ZeroValue:          7,
		NilValue:           &str,
		IgnoredField:       "Ignored",
		IgnoredFieldTwo:    "Ignored Two",
		IgnoredFieldByFunc: "Ignored By Func",
		IgnoredByTag:       "Ignored By Tag",
	}

	n := Something{
		Number:             6,
		Text:               "Old Text",
		NewText:            "New Text",
		ZeroValue:          0,
		NilValue:           nil,
		IgnoredField:       "Diff Ignored",
		IgnoredFieldTwo:    "Diff Ignored Two",
		IgnoredFieldByFunc: "Diff Ignored By Func",
		IgnoredByTag:       "Diff Ignored By Tag",
	}

	// The patcher.LoadDiff function will apply the changes from n to s.
	if err := patcher.LoadDiff(&s, &n,
		patcher.WithIncludeZeroValues(true),
		patcher.WithIncludeNilValues(true),
		patcher.WithIgnoredFields("ignoredField", "IgNoReDfIeLdTwO"),
		patcher.WithIgnoredFieldsFunc(func(field *reflect.StructField) bool {
			return strings.EqualFold(field.Name, "ignoredfieldbyfunc")
		}),
	); err != nil {
		panic(err)
	}

	// Output:
	// 6
	// Old Text
	// PrePopulated
	// New Text
	// 0
	// <nil>
	// Ignored
	// Ignored Two
	// Ignored By Func
	// Ignored By Tag
	fmt.Println(s.Number)
	fmt.Println(s.Text)
	fmt.Println(s.PrePopulated)
	fmt.Println(s.NewText)
	fmt.Println(s.ZeroValue)
	fmt.Println(s.NilValue)
	fmt.Println(s.IgnoredField)
	fmt.Println(s.IgnoredFieldTwo)
	fmt.Println(s.IgnoredFieldByFunc)
	fmt.Println(s.IgnoredByTag)
}
