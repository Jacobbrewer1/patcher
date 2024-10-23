#!/usr/bin/env bash

# Remove all mock files in the repo (These are identified with the "mock_" prefix)
find . -name "mock_*" -exec rm -rf {} \;

# Generate mock files for all interfaces in the `./pkg` directory
go generate ./...
