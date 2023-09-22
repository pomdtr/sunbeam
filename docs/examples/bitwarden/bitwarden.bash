#!/bin/bash

if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

if ! command -v bw &> /dev/null; then
    echo "bw is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Bitwarden Vault",
        commands: [
            {
                name: "list-passwords",
                title: "List Passwords",
                mode: "page",
                params: [
                    {name: "session", type: "string", optional: true, description: "session token"}
                ]
            }
        ]
    }'
    exit 0
fi

if [ "$1" = "list-passwords" ]; then
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
    }' | jq -s '{type: "list", items: .}'
fi
