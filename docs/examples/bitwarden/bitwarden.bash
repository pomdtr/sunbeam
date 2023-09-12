#!/bin/bash

if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Bitwarden Vault",
        commands: [
            {
                name: "list",
                title: "List Passwords",
                mode: "filter",
            }
        ]
    }'
    exit 0
fi

if [ "$1" = "list" ]; then
    bw --nointeraction list items --session "$BW_SESSION" | jq '.[] | {
        title: .name,
        subtitle: (.login.username // ""),
        actions: [
            {
                type: "copy",
                title: "Copy Password",
                exit: true,
                text: (.login.password // ""),
            },
            {
                type: "copy",
                title: "Copy Username",
                exit: true,
                text: (.login.username // ""),
                key: "l"
            }
        ]
    }' | jq -s '{items: .}'
fi
