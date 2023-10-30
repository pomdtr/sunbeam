#!/bin/sh
# shellcheck disable=SC2016

# This script is an example of how to use tldr with sunbeam
set -eu

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
    sunbeam query -n '
{
    title: "Browse TLDR Pages",
    description: "Browse TLDR Pages",
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "page", title: "Search Pages" },
        { name: "view", mode: "page", title: "View page", params: [{ name: "page", type: "string", required: true, title: "page to show" }] }
    ]
}'
exit 0
fi

# check if tldr is installed
if ! command -v tldr >/dev/null 2>&1; then
    echo "tldr is not installed. Please install it from https://dbrgn.github.io/tealdeer/installing.html"
    exit 1
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list" ]; then
    tldr --list | sunbeam query -R '{
        title: .,
        actions: [
            {title: "View Page", type: "run", command: "view", params: {page: .}},
            {title: "Copy Command", key: "c", type: "copy", text: ., exit: true }
        ]
    }' | sunbeam query -s '{ type: "list", items: . }'
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(echo "$1" | sunbeam query -r '.params.page')
    tldr --color=always "$PAGE" | sunbeam query --arg page="$PAGE" -sR '{
            type: "detail", highlight: "ansi", text: ., actions: [
                {title: "Copy Page", type: "copy", text: ., exit: true},
                {title: "Copy Command", key: "c", type: "copy", text: $page, exit: true}
            ]
        }'
fi
