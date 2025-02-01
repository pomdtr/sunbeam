#!/usr/bin/env bash

set -euo pipefail

EXTENSIONS_DIR="${SUNBEAM_EXTENSIONS_DIR:-$HOME/.config/sunbeam/extensions}"

if [ $# -eq 0 ]; then
  jq -n '{
    title: "Sunbeam",
    root: [
      { title: "Search Extensions", type: "run", command: "search-extensions" }
    ],
    commands: [
      { name: "search-extensions", mode: "filter" }
    ]
  }'
  exit 0
fi

COMMAND=$1

if [ "$COMMAND" = "search-extensions" ]; then
    ls "$EXTENSIONS_DIR" | jq --arg dir "$EXTENSIONS_DIR" -R '{
        title: .,
        accessories: [
            "\($dir)/\(.)"
        ],
        actions: [
            { title: "Edit Extension", type: "edit", path: "\($dir)/\(.)" },
            { title: "Copy Path", type: "copy", text: "\($dir)/\(.)" }
            
        ]
    }' | jq -s '{ items: . }'
fi