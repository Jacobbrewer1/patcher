#!/usr/bin/env bash

# Remove all mock files in the repo (These are identified with the "mock_" prefix.
files=$(find ./)

for file in $files; do
  echo "$file"
done
