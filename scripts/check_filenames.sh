#!/usr/bin/env bash

# Find all Go files in the project directory and its subdirectories
for file in $(find . -name "*.go"); do

  # Check if the file name contains uppercase letters
  if [[ "$file" =~ [A-Z] ]]; then
    echo "Error: $file contains uppercase letters. All Go files in the project must use snake_case"
    exit 1
  fi
done

echo "All Go files in the project use lowercase letters"
exit 0
