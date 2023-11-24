# Google Search

This scripts allows you to search Google from Sunbeam.

```bash
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Google Search",
        description: "Search the web with Google",
        root: ["search"],
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
COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "search" ]; then
    # Get the query from the user
    QUERY=$(echo "$1" | sunbeam query -r '.query')

    # If the query is empty, show an help message
    if [ "$QUERY" = "null" ]; then
        sunbeam query -n '{
            emptyText: "Type something to search",
        }'
        exit 0
    fi

    # urlencode the query
    QUERY=$(echo "$QUERY" | sunbeam query -rR '@uri')
    sunbeam fetch "https://suggestqueries.google.com/complete/search?client=firefox&q=$QUERY" | sunbeam query '.[1] | {
        items: map({
            title: .,
            actions: [
                { title: "Search", type: "open", url: "https://www.google.com/search?q=\(.)", exit: true }
            ]
        })
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
```
