#!/bin/bash

# check for dependencies

if ! command -v jq &> /dev/null; then
    echo "jq could not be found"
    exit 1
fi

if ! command -v bkt &> /dev/null; then
    echo "bkt could not be found"
    exit 1
fi

if ! command -v bw &> /dev/null; then
    echo "bw could not be found"
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Bitwarden Vault",
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

COMMAND=$(jq -r '.command' <<< "$1")
if [ "$COMMAND" = "list-passwords" ]; then
    bkt --ttl 24h --stale 60s -- bw --nointeraction list items --session "$BW_SESSION" | jq '.[] | {
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
    }' | jq -s '{type: "list", items: .}'
fi
