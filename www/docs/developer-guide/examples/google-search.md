# Google Search

This scripts allows you to search Google from Sunbeam.

Refer to the devdocs extension for a more complete example of writing an extension as a shell script.
This example demonstrates how to write an extension that uses the search mode.

Each time the user types a character, the script is called with the query as argument.
The emptyText field is used to display a message when no items are shown in the list.

```bash
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Google Search",
        description: "Search the web with Google",
        commands: [
            {
                name: "search",
                mode: "search",
                title: "Search",
            }
        ]
    }'
    exit 0
fi

# since the command is a search, the script is called with the query as argument every time the user types a character
COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "search" ]; then
    # Get the query from the user
    QUERY=$(echo "$1" | jq -r '.query')

    # If the query is empty, show an help message
    if [ "$QUERY" = "null" ]; then
        jq -n '{
            emptyText: "Type something to search",
        }'
        exit 0
    fi

    # urlencode the query
    curl -G "https://suggestqueries.google.com/complete/search" -d "q=$QUERY" | jq '.[1] | {
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", url: "https://www.google.com/search?q=\(.)"}
            ]
        })
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
```
