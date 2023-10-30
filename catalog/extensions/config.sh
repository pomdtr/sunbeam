#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Edit Config Files",
        description: "Edit your favorite config files",
        commands: [
            {
                name: "fish",
                title: "Edit Fish Config",
                mode: "tty"
            },
            {
                name: "hyper",
                title: "Edit Hyper Config",
                mode: "tty"
            },
            {
                name: "sunbeam",
                title: "Edit Sunbeam Config",
                mode: "tty"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')

if [ "$COMMAND" = "fish" ]; then
    sunbeam edit ~/.config/fish/config.fish
elif [ "$COMMAND" = "hyper" ]; then
    sunbeam edit ~/.hyper.js
elif [ "$COMMAND" = "sunbeam" ]; then
    sunbeam edit ~/.config/sunbeam/config.json
fi
