#!/usr/bin/env bash

# Check formatting
res=$(gofmt -l .)
if [[ "$res" != "" ]]; then
  echo "Error: the following files are not formatted correctly:"
  for file in $res; do
    echo " - $file"
  done
  echo "Run 'gofmt -s -w .' to fix this error"
  exit 1
fi

echo "All files in the project are formatted correctly"
exit 0
