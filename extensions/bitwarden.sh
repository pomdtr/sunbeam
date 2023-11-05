#!/bin/bash

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Bitwarden Vault",
        description: "List your Bitwarden passwords",
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

# check for dependencies
if ! command -v bkt &> /dev/null; then
    echo "bkt could not be found"
    exit 1
fi

if ! command -v bw &> /dev/null; then
    echo "bw could not be found"
    exit 1
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list-passwords" ]; then
    bkt --ttl 24h --stale 1h -- bw --nointeraction list items --session "$BW_SESSION" | sunbeam query 'map({
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
