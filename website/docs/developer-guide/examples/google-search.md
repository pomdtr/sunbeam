# Google Search

This scripts allows you to search Google from Sunbeam. The list of suggestions is refreshed every time the user types a character, thanks to the `dynamic` property.

```bash
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Google Search",
        description: "Search the web with Google",
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
    # Get the query from the user
    QUERY=$(echo "$1" | sunbeam query -r '.query')

    # If the query is empty, show an help message
    if [ "$QUERY" = "null" ]; then
        # the dynamic property tells Sunbeam to rerun the command every time the user types a character
        sunbeam query -n '{
            dynamic: true,
            emptyText: "Type something to search",
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
```
