#!/bin/bash

set -e

for filename in examples/*/sunbeam.yml; do
  echo "Validating $filename"
  sunbeam lint "$filename"
done
