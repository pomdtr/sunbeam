#!/usr/bin/env bash

# This script is an example of how to use tldr with sunbeam
set -euo pipefail

# check if tldr is installed
if ! command -v tldr &> /dev/null; then
    echo "tldr is not installed"
    exit 1
fi


# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
    jq -n '
{
    title: "TLDR Pages",
    # each command can be called through the cli
    commands: [
        { name: "list", mode: "view", title: "List TLDR pages" },
        { name: "view", mode: "view", title: "View TLDR page", params: [{ name: "page", type: "string", description: "page to show" }] }
    ]
}'
exit 0
fi

# extract command name
COMMAND=$(jq -r '.command' <<< "$1")
# handle commands
if [ "$COMMAND" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {title: "View Page", onAction: { type: "run", command: "view", params: {page: .}}},
            {title: "Copy Command", key: "c", onAction: { type: "copy", text: ., exit: true }}
        ]
    }' | jq -s '{
            type: "list",
            items: .
        }'
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(jq -r '.params.page' <<< "$1")
    # shellcheck disable=SC2016
    tldr --raw "$PAGE" | jq --arg page "$PAGE" -sR '{
            type: "detail", text: ., language: "markdown", actions: [
                {title: "Copy Page", onAction: {type: "copy", text: ., exit: true}},
                {title: "Copy Command", onAction: {type: "copy", text: $page, exit: true}}
            ]
        }'
fi
