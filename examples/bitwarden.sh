#!/bin/bash

bw list items | jq -cM '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy",
            title: "Copy Password",
            shortcut: "enter",
            content: .login.password
        }
    ]
}'
