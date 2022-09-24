#!/bin/sh

# @sunbeam.schemaVersion 1
# @sunbeam.title Remote Script
# @sunbeam.mode command
# @sunbeam.packageName Remote

curl \
    -X POST \
    -H "Content-type: application/json" \
    --data-binary @- \
    http://localhost:8000
