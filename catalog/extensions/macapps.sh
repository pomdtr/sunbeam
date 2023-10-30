#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Mac Apps",
        description: "Open your favorite apps",
        commands: [
            {
                name: "list",
                title: "List All Apps",
                mode: "page"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
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
    }' | sunbeam query -s '{ type: "list", items: . }'
fi
