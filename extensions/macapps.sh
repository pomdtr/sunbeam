#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Mac Apps",
        platforms: ["macos"],
        description: "Open your favorite apps",
        root: [
            { command: "list" }
        ],
        commands: [
            {
                name: "list",
                title: "List All Apps",
                mode: "list"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "list" ]; then
    mdfind kMDItemContentTypeTree=com.apple.application-bundle -onlyin / | sunbeam query -R '{
        title: (split("/")[-1] | split(".app")[0]),
        accessories: [.],
        actions: [{
            title: "Open",
            type: "open",
            target: .,
            exit: true
        }]
    }' | sunbeam query -s '{  items: . }'
fi
