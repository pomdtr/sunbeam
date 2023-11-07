#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Google Search",
        commands: [
            {
                name: "search",
                mode: "list",
                title: "Search",
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "search" ]; then
    QUERY=$(echo "$1" | sunbeam query -r '.query')
    if [ "$QUERY" = "null" ]; then
        sunbeam query -n '{
            dynamic: true,
            emptyText: "Type something to search",
        }'
        exit 0
    fi
    sunbeam fetch "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | sunbeam query '.[1] | {
        dynamic: true,
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", target: "https://www.google.com/search?q=\(.)" }
            ]
        })
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
