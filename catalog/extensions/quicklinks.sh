#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Quick Links",
        description: "Open your favorite websites",
        commands: [
            {
                name: "google",
                title: "Open Google",
                mode: "silent"
            },
            {
                name: "workflowy-sunbeam",
                title: "Open Sunbeam Tree",
                mode: "silent"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "google" ]; then
    sunbeam open https://google.com
elif [ "$COMMAND" = "workflowy-sunbeam" ]; then
    sunbeam open https://workflowy.com/#/56ff341c7433
fi
