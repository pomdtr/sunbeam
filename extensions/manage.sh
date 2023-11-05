#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Manage Extensions",
        description: "Manage Sunbeam Extensions",
        commands: [
            {
                name: "list-extensions",
                title: "List Extensions",
                mode: "list"
            },
            {
                name: "remove-extension",
                title: "Remove Extension",
                mode: "silent",
                params: [
                    {
                        name: "name",
                        description: "Extension Name",
                        required: true,
                        type: "string"
                    }
                ]
            },
            {
                name: "edit-extension",
                title: "Edit Extension",
                mode: "tty",
                params: [
                    {
                        name: "name",
                        description: "Extension Name",
                        required: true,
                        type: "string"
                    }
                ]
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list-extensions" ]; then
    sunbeam extension list --json | sunbeam query '{

        items: map({
            title: .alias,
            subtitle: .manifest.title,
            accessories: [.type],
            actions: [
                {
                    "title": "Edit",
                    "type": "run",
                    "command": "edit-extension",
                    "params": {
                        "name": .alias
                    },
                    reload: true
                },
                {
                    "title": "Remove",
                    "type": "run",
                    "command": "remove-extension",
                    "params": {
                        "name": .alias
                    },
                    reload: true
                }
            ]
        })
    }'
elif [ "$COMMAND" = "remove-extension" ]; then
    NAME=$(echo "$1" | sunbeam query -r '.params.name')
    sunbeam extension remove "$NAME"
elif [ "$COMMAND" = "edit-extension" ]; then
    NAME=$(echo "$1" | sunbeam query -r '.params.name')
    sunbeam extension edit "$NAME"
fi
