package inserter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBatch(t *testing.T) {
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

	b := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources))

	fmt.Println(b.Fields())
	fmt.Println(b.Args())
}

func TestNewBatchSQL(t *testing.T) {
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

	sql, args, err := NewBatch(WithTable("temp"), WithTagName("db"), WithResources(resources)).sqlGen()
	require.NoError(t, err)

	fmt.Println(sql)
	fmt.Println(args)
}
