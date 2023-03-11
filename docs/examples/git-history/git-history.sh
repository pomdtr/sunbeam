#!/bin/bash

set -euo pipefail

git log | jc --git-log | sunbeam query '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Copy Commit Hash", text: .commit, type: "copy" }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    items: .
}'
