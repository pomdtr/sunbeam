#!/bin/bash

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Bitwarden Vault",
        description: "Search your Bitwarden passwords",
        env: [
            { name: "BW_SESSION", description: "Bitwarden Session Token", required: true }
        ],
        requirements: [
            { name: "bw", link: "https://bitwarden.com/help/article/cli/" }
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

COMMAND=$(echo "$1" | jq -r '.command')
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
