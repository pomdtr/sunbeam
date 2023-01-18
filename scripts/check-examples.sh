#!/bin/bash

set -e

for filename in examples/*/sunbeam.yml; do
  echo "Checking $filename..."
  sunbeam check "$filename"
done
