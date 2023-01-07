#!/bin/bash

set -euo pipefail

git -C "$1" log | jc --git-log | sunbeam query '.[] | {
    title: (.message | split("\n") | .[0]),
    subtitle: .commit[:7],
    accessories: [.author, .date],
    actions: [
        { title: "Copy Commit Hash", text: .commit, type: "copy-text" }
    ]
}'
