#!/bin/bash

set -euo pipefail

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
