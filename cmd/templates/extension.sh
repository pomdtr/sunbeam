#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Example Extension",
        description: "Example extension",
        commands: [
            {
                name: "hello",
                title: "Hello",
                mode: "detail"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r .command)
if [ "$COMMAND" = "hello" ]; then
    sunbeam query -n '{
        text: "Hello, world!"
    }'
fi
