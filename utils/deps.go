//go:build deps

package utils

import (
	_ "github.com/vektra/mockery/v2" // Mockery is a tool for generating mocks for interfaces in Go. This prevents the tool from being removed when running go mod tidy.
)
