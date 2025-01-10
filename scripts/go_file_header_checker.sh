#!/bin/bash

# Do not run this file directly. It is intended to be run as a pre-commit hook.
# To install this hook, run 'pre-commit install' in the root of the repository.

echo "Running Go File Header Checker..."

go_files=$(git ls-files '*.go')
all_files_valid=true

for file in $go_files; do
  dir=$(dirname "$file")
  filename=$(basename "$file")
  first_line=$(head -n 1 "$file")
  expected_comment="// $dir/$filename"

  if [[ "$first_line" != "$expected_comment" ]]; then
    echo "File $file does not start with the correct comment."
    echo "Expected: $expected_comment"
    echo "Found: $first_line"
    all_files_valid=false
  fi
done

if [ "$all_files_valid" = false ]; then
  echo "Please fix the header comments in the listed files before committing."
  exit 1
fi

echo "All Go files have correct header comments."
exit 0

