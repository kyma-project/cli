#!/bin/bash

echo "### Verify code standard output usage ###"
code_std_out_usage=$(grep -r -E 'fmt\.Print|os\.Stdout|os\.Stderr' ./internal | grep --invert-match '^./internal/out')
if [ -n "$code_std_out_usage" ]; then
  echo "Found usage of os.Stdout, os.Stderr or fmt.Print in code:"
  echo "$code_std_out_usage"
  echo ""
  echo "Please use the internal/out package for output handling instead."
  exit 1
fi

echo "OK"
echo ""
