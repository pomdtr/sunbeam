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
    items: [{ command: "list"}],
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "list", title: "Search Pages" },
        { name: "view", mode: "detail", title: "View page", params: [{ name: "page", type: "text", required: true, title: "page to show" }] },
        { name: "update", mode: "silent", title: "Update cache" }
    ]
}'
exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "list" ]; then
    tldr --list | sunbeam query -R '{
        title: .,
        actions: [
            {title: "View Page", type: "run", command: "view", params: {page: .}},
            {title: "Update Cache", key: "r", type: "run", command: "update", reload: true }
        ]
    }' | sunbeam query -s '{ items: . }'
elif [ "$COMMAND" = "update" ]; then
    tldr --update
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(echo "$1" | sunbeam query -r '.params.page')
    tldr --color=always "$PAGE" | sunbeam query --arg page="$PAGE" -sR '{
            format: "ansi", text: ., actions: [
                {title: "Copy Page", type: "copy", text: ., exit: true}
            ]
        }'
fi
