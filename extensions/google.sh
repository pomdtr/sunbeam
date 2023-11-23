#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Google Search",
        description: "Search Google",
        root: [ "search" ],
        commands: [
            {
                name: "search",
                mode: "search",
                title: "Search Google",
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
            emptyText: "Type anything to search",
        }'
        exit 0
    fi
    # urlencode the query
    QUERY=$(echo "$QUERY" | sunbeam query -rR '@uri')
    sunbeam fetch "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | sunbeam query '.[1] | {
        dynamic: true,
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", target: "https://www.google.com/search?q=\(.)", exit: true }
            ]
        })
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
