#!/usr/bin/env bash

set -euo pipefail

# check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '
{
    title: "TLDR Pages",
    commands: [
        {
            name: "list",
            title: "List TLDR pages",
            mode: "filter",
        },
        {
            name: "view",
            title: "View TLDR page",
            mode: "detail",
            params: [
                {
                    name: "command",
                    type: "string",
                }
            ]
        }
    ]
}'
exit 0
fi

eval "$(sunbeam parse bash)"
if [ "$1" = "list" ]; then
    tldr --list | jq -R '{
        title: .,
        actions: [
            {type: "run", title: "View", command: "view", params: {command: .}}
        ]
    }' | jq -s '{items: .}'
elif [ "$1" = "view" ]; then
    tldr --raw "$COMMAND" | jq -sR '{text: ., language: "markdown", actions: [{type: "copy", title: "Copy", text: ., exit: true}]}'
fi
