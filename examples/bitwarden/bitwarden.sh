#!/bin/bash

set -euo pipefail

BW_SESSION=$(sunbeam kv get session || true)

if ! bw unlock --check --session "$BW_SESSION" &>/dev/null; then
    sunbeam query -n '{
        type: "detail",
        preview: "Vault is locked",
        actions: [
            {
                type: "run-command",
                title: "Unlock Vault",
                command: "unlock-vault",
                onSuccess: "reload-page",
                with: {
                    password: {
                        type: "password",
                        title: "Master Password"
                    }
                }
            }
        ]
    }'
    exit 0
fi

bw --nointeraction list items --session "$BW_SESSION" | sunbeam query '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy-text",
            title: "Copy Password",
            shortcut: "enter",
            text: .login.password
        },
        {
            type: "copy-text",
            title: "Copy Login",
            shortcut: "ctrl+y",
            text: "\(.login)"
        },
        {
            "type": "run-command",
            "title": "Lock Vault",
            "shortcut": "ctrl+l",
            "onSuccess": "reload-page",
            "command": "lock-vault"
        }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    items: .
}'
