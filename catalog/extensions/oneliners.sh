#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Oneliners",
        description: "Run oneliners from sunbeam",
        commands: [
            {
                name: "run",
                title: "htop",
                mode: "tty",
                params: [
                    {
                        name: "oneliner",
                        type: "string",
                        description: "oneliner to run",
                        required: true
                    }
                ]
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "run" ]; then
    ONELINER=$(echo "$1" | jq -r '.params.oneliner')
    sh -c "$ONELINER"
fi
