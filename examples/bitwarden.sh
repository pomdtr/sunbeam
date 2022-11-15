#!/bin/bash

bw list items | sunbeam jq '.[] | {
    title: .name,
    subtitle: .login.username,
    actions: [
        {
            type: "copy-text",
            title: "Copy Password",
            shortcut: "enter",
            text: .login.password
        }
    ]
}'
