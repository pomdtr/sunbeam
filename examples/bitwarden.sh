#!/bin/bash

bw list items | sunbeam query '.[] | {
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
            title: "Copy Password",
            shortcut: "ctrl+y",
            text: "\(.login)"
        }
    ]
}'
