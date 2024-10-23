#!/usr/bin/env bash

# Remove all mock files in the repo (These are identified with the "mock_" prefix)
files=$(find . -name "mock_*.go")
for file in $files; do
  echo "Removing $file"
  rm -rf "$file"
done

# Generate mock files for all interfaces in the `./pkg` directory
go generate ./...
