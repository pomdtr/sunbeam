#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Quick Links",
        description: "Open your favorite websites",
        commands: [
            {
                name: "open",
                title: "Open link",
                mode: "silent",
                params: [
                    {
                        name: "url",
                        description: "URL",
                        type: "string",
                        required: true
                    }
                ],
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "open" ]; then
    URL=$(echo "$1" | jq -r '.params.url')
    sunbeam open "$URL"
fi
