package main

import (
	"fmt"

	"github.com/jacobbrewer1/patcher"
)

type ExampleFilter struct {
	Field string
	Value any
}

func (e ExampleFilter) Where() (string, []any) {
	return fmt.Sprintf("%s = ?", e.Field), []any{e.Value}
}

func main() {
	// Create example filters
	filters := []ExampleFilter{
		{Field: "name", Value: "John"},
		{Field: "age", Value: 30},
	}

	mf := patcher.NewMultiFilter()

	// Append each filter to the WHERE clause
	for _, filter := range filters {
		mf.Add(filter)
	}

	// Print the WHERE clause
	sql, args := mf.Where()
	fmt.Printf("WHERE SQL:\n%s\n", sql)
	fmt.Printf("WHERE Args: %v\n", args)
}
