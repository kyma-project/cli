#!/bin/bash

echo "### Verify git tree ###"
local_changes=$(git status --porcelain)
if [ -n "$local_changes" ]; then
  echo "There are local changes in the git tree. Please commit or stash them before running the verification."
  echo "$local_changes"
  exit 1
fi

echo "OK"
echo ""

echo "### Verify documentation ###"
make docs
docs_changes=$(git status --porcelain)
if [ -n "$docs_changes" ]; then
  echo "Documentation is out of date. Please run 'make docs' and commit the changes."
  echo "$docs_changes"
  exit 1
fi

echo "OK"
echo ""

echo "### Verify code standard output usage ###"
code_std_out_usage=$(grep -r -e 'fmt.Print' -e 'os.Stdout' -e 'os.Stderr' ./internal | grep --invert-match '^./internal/out')
if [ -n "$code_std_out_usage" ]; then
  echo "Found usage of os.Stdout, os.Stderr or fmt.Print:"
  echo "$code_std_out_usage"
  exit 1
fi

echo "OK"
echo ""
