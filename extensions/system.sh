#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "System",
        description: "Control your system",
        platforms: ["macos"],
        commands: [
            {
                name: "toggle-dark-mode",
                title: "Toggle Dark Mode",
                mode: "silent"
            },
            {
                name: "lock-screen",
                title: "Lock Screen",
                mode: "silent"
            },
            {
                name: "empty-trash",
                title: "Empty Trash",
                mode: "silent"
            },
            {
                name: "open-trash",
                title: "Open Trash",
                mode: "silent"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "toggle-dark-mode" ]; then
    osascript -e 'tell app "System Events" to tell appearance preferences to set dark mode to not dark mode'
elif [ "$COMMAND" = "lock-screen" ]; then
    osascript -e 'tell application "System Events" to keystroke "q" using {command down,control down}'
elif [ "$COMMAND" = "empty-trash" ]; then
    osascript -e 'tell application "Finder" to empty trash'
elif [ "$COMMAND" = "open-trash" ]; then
    osascript -e 'tell application "Finder" to open trash'
fi
