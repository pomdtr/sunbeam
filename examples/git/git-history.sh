#!/bin/bash

set -euo pipefail

git -C "$1" log | jc --git-log | jq '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Copy Commit Hash", text: .commit, type: "copy-text" }
    ]
}' | jq --slurp '{
    type: "list",
    list: {items: .}
}'
