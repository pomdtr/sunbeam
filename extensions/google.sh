#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Google Search",
        description: "Search Google",
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

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "search" ]; then
    QUERY=$(echo "$1" | jq -r '.query')
    if [ "$QUERY" = "null" ]; then
        jq -n '{
            emptyText: "Type anything to search",
        }'
        exit 0
    fi
    # urlencode the query
    QUERY=$(echo "$QUERY" | jq -rR '@uri')
    curl "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | jq '.[1] | {
        dynamic: true,
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", url: "https://www.google.com/search?q=\(.)" }
            ]
        })
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
