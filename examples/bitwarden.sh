#!/bin/bash

bw list items | sunbeam jq '.[] | {
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
