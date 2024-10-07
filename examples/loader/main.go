package main

import (
	"fmt"

	"github.com/Jacobbrewer1/patcher"
)

type Something struct {
	Number       int
	Text         string
	PrePopulated string
	NewText      string
}

func main() {
	s := Something{
		Number:       5,
		Text:         "Hello",
		PrePopulated: "PrePopulated",
	}

	n := Something{
		Number:  6,
		Text:    "Old Text",
		NewText: "New Text",
	}

	// The patcher.LoadDiff function will apply the changes from n to s.
	if err := patcher.LoadDiff(&s, &n); err != nil {
		panic(err)
	}

	// Output:
	// 6
	// Old Text
	// PrePopulated
	// New Text
	fmt.Println(s.Number)
	fmt.Println(s.Text)
	fmt.Println(s.PrePopulated)
	fmt.Println(s.NewText)
}
