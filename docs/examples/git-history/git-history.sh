#!/bin/bash

set -euo pipefail

git log | jc --git-log | sunbeam query '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Checkout Commit", type: "run", command: "git checkout \(.commit)" },
        { title: "Copy Commit Hash", text: .commit, type: "copy", shortcut: "ctrl+h" },
        { title: "Copy Message", text: .message, type: "copy", shortcut: "ctrl+m" }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    items: .
}'
