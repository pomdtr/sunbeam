#!/bin/bash

# check for dependencies

if ! command -v bkt &> /dev/null; then
    echo "bkt could not be found"
    exit 1
fi

if ! command -v bw &> /dev/null; then
    echo "bw could not be found"
    exit 1
fi

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Bitwarden Vault",
        root: "list-passwords",
        commands: [
            {
                name: "list-passwords",
                title: "List Passwords",
                mode: "view",
                params: [
                    {name: "session", type: "string", optional: true, description: "session token"}
                ]
            }
        ]
    }'
    exit 0
fi

COMMAND=$(sunbeam query -r '.command' <<< "$1")
if [ "$COMMAND" = "list-passwords" ]; then
    bkt --ttl 24h --stale 1h -- bw --nointeraction list items --session "$BW_SESSION" | sunbeam query '.[] | {
        title: .name,
        subtitle: (.login.username // ""),
        actions: [
            {
                title: "Copy Password",
                onAction: {
                    type: "copy",
                    text: (.login.password // ""),
                    exit: true
                }
            },
            {
                title: "Copy Username",
                key: "l",
                onAction: {
                    type: "copy",
                    text: (.login.username // ""),
                    exit: true
                }
            }
        ]
    }' | sunbeam query -s '{type: "list", items: .}'
fi
