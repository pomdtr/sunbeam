#!/bin/bash

bw list items | sunbeam jq '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy-content",
            title: "Copy Password",
            shortcut: "enter",
            content: .login.password
        }
    ]
}'
