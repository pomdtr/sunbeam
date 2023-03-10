#!/bin/bash

set -euo pipefail

git log | jc --git-log | jq '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Copy Commit Hash", text: .commit, type: "copy" }
    ]
}' | jq --slurp '{
    type: "list",
    items: .
}'
