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
    root: [
        { title: "List Pages", type: "run", command: "list" }
    ],
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "filter", description: "Search Pages" },
        { name: "view", mode: "detail", description: "View page", params: [{ name: "page", type: "string", description: "Page to show" }] }
    ]
}'
exit 0
fi

# check if tldr is installed
if ! [ -x "$(command -v tldr)" ]; then
    echo "tldr is not installed. Please install it." >&2
    exit 1
fi

if [ "$1" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {title: "View Page", type: "run", command: "view", params: {page: .}},
        ]
    }' | jq -s '{ items: . }'
elif [ "$1" = "view" ]; then
    PAGE=$(cat | jq -r '.page')
    tldr --raw "$PAGE" | jq -sR '{
            markdown: ., actions: [
                {title: "Copy Page", type: "copy", text: .}
            ]
        }'
fi
