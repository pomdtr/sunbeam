#!/bin/sh

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Bitwarden Vault",
        description: "Search your Bitwarden passwords",
        preferences: [
            {
                name: "session",
                title: "Bitwarden Session",
                type: "text",
                required: true
            }
        ],
        items: [
            { command: "list-passwords" }
        ],
        commands: [
            {
                name: "list-passwords",
                title: "List Passwords",
                mode: "list"
            }
        ]
    }'
    exit 0
fi

BW_SESSION=$(echo "$1" | sunbeam query -r '.preferences.session')
if [ "$BW_SESSION" = "null" ]; then
    echo "Bitwarden session not found. Please set it in your config." >&2
    exit 1
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "list-passwords" ]; then
    bw --nointeraction list items --session "$BW_SESSION" | sunbeam query 'map({
        title: .name,
        subtitle: (.login.username // ""),
        actions: [
            {
                title: "Copy Password",
                type: "copy",
                text: (.login.password // ""),
                exit: true
            },
            {
                title: "Copy Username",
                key: "l",
                type: "copy",
                text: (.login.username // ""),
                exit: true
            }
        ]
    }) | { items: .}'
fi
