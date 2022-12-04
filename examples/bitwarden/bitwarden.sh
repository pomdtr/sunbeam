#!/bin/bash

bw list items | sunbeam query '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copyText",
            title: "Copy Password",
            shortcut: "enter",
            text: .login.password
        },
        {
            type: "copyText",
            title: "Copy Login",
            shortcut: "ctrl+y",
            text: "\(.login)"
        }
    ]
}'
