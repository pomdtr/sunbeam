#!/bin/sh
# shellcheck disable=SC2016

# This script is an example of how to use tldr with sunbeam
set -eu

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
    jq -n '
{
    title: "Browse TLDR Pages",
    description: "Browse TLDR Pages",
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "filter", title: "Search Pages" },
        { name: "view", mode: "detail", hidden: true, title: "View page", params: [{ name: "page", type: "string", title: "Page to show" }] },
        { name: "update", mode: "silent", title: "Update cache" }
    ]
}'
exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {title: "View Page", type: "run", command: "view", params: {page: .}},
            {title: "Update Cache", key: "r", type: "run", command: "update", reload: true}
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "update" ]; then
    tldr --update
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(echo "$1" | jq -r '.params.page')
    tldr --raw "$PAGE" | jq --arg page "$PAGE" -sR '{
            markdown: ., actions: [
                {title: "Copy Page", type: "copy", copy: {text: ., exit: true}}
            ]
        }'
fi
