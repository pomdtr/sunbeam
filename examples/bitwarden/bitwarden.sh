#!/bin/bash

set -eo pipefail

if ! bw unlock --check --session "$BW_SESSION" >/dev/null 2>&1; then
    echo "The vault is locked, update your session token." >&2
    exit 1
fi

bw --nointeraction list items --session "$BW_SESSION" | sunbeam query '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy-text",
            title: "Copy Password",
            text: .login.password
        },
        {
            type: "copy-text",
            title: "Copy Login",
            text: "\(.login)"
        },
        {
            "type": "run-command",
            "title": "Lock Vault",
            "onSuccess": "reload-page",
            "command": "lock-vault"
        }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    items: .
}'
