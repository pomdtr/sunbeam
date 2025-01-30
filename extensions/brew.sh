#!/bin/sh

# check if jq is installed
if ! [ -x "$(command -v jq)" ]; then
    echo "jq is not installed. Please install it." >&2
    exit 1
fi

# check if brew is installed
if ! [ -x "$(command -v brew)" ]; then
    echo "brew is not installed. Please install it." >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Brew",
        root: [
            { title: "List Installed Packages", type: "run", command: "list" }
        ],
        commands: [
            {
                name: "list",
                description: "List installed packages",
                mode: "filter",
            },
            {
                name: "info",
                mode: "detail",
                description: "Show info about a package",
                params: [
                    { name: "package", type: "string" }
                ]
            }
        ]
    }'
    exit 0
fi

if [ "$1" = "list" ]; then
    brew list | jq -R '{
        title: .,
        actions: [
            {
                title: "Show Info",
                type: "run",
                command: "info",
                params: { package: . }
            }
        ]
    }' | jq -s '{ items: . }'
elif [ "$1" = "info" ]; then
    PACKAGE=$(cat | jq -r '.package')
    brew info "$PACKAGE" | jq -sR '{ text: . }'
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
