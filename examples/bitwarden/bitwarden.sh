#!/bin/bash

set -euo pipefail

bw --nointeraction list items | sunbeam query '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy",
            title: "Copy Password",
            shortcut: "enter",
            text: .login.password
        },
        {
            type: "copy",
            title: "Copy Login",
            shortcut: "ctrl+y",
            text: "\(.login)"
        }
    ]
}'
