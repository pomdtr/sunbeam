#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Google Search",
        description: "Search Google",
        actions: [
            { title: "Search Google", type: "run", command: "search" }
        ],
        commands: [
            {
                name: "search",
                mode: "search",
                description: "Search Google",
                params: [
                    { name: "query", title: "Query", type: "string" }
                ]
            }
        ]
    }'
    exit 0
fi


COMMAND=$1
QUERY=""
while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--query" ]]; then
        QUERY="$2"
        break
    fi
    shift
done

if [ "$COMMAND" = "search" ]; then
    # urlencode the query
    QUERY=$(echo "$QUERY" | jq -rR '@uri')
    curl "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | jq '.[1] | {
        dynamic: true,
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", target: "https://www.google.com/search?q=\(.)" }
            ]
        })
    }'
else
    echo "Unknown command: $1" >&2
    exit 1
fi
