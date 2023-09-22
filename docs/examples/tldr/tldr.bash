#!/usr/bin/env bash

# This script is an example of how to use tldr with sunbeam
set -euo pipefail

# check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

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
        { name: "list", mode: "page", title: "List TLDR pages" },
        { name: "view", mode: "page", title: "View TLDR page", params: [{ name: "page", type: "string", description: "page to show" }] }
    ]
}'
exit 0
fi

# read input from stdin
INPUT=$(cat)
COMMAND=$1

# handle commands
if [ "$COMMAND" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {type: "run", title: "View Page", command: { name: "view", params: {page: .}}}
        ]
    }' | jq -s '{
            type: "list",
            items: .
        }'
elif [ "$COMMAND" = "view" ]; then
    PAGE=$(jq -r '.params.page' <<< "$INPUT")
    tldr --raw "$PAGE" | jq -sR '{
            type: "detail", text: ., language: "markdown", actions: [{type: "copy", title: "Copy", exit: true, text: .}]
        }'
fi
