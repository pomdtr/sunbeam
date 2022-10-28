#!/bin/bash

bw list items | jq -cM '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy",
            shortcut: "enter",
            content: .login.password
        }
    ]
}'
