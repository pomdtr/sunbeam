#!/bin/bash

set -euo pipefail

git log | jc --git-log | sunbeam query '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Checkout Commit", type: "run-command", command: "git checkout \(.commit)" },
        { title: "Copy Commit Hash", text: .commit, type: "copy", key: "h" },
        { title: "Copy Message", text: .message, type: "copy", key: "m" }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    items: .
}'
