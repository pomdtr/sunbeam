#!/usr/bin/env bash

set -eo pipefail

if [ -n "$1" ]; then
    ENDPOINT=/users/$1/repos
else
    ENDPOINT=/user/repos
fi

gh api "$ENDPOINT" --paginate --cache 3h --jq '.[] |
    {
        id: (.id | tostring),
        title: .name,
        preview: (.description // "No description"),
        accessories: [
            "\(.stargazers_count) *"
        ],
        actions: [
            {type: "open", url: .html_url},
            {
                type: "run-script",
                script: "list-pull-requests",
                title: "List Pull Requests",
                shortcut: "ctrl+p",
                with: {repository: .full_name}
            },
            {
                type: "run-script",
                script: "view-readme",
                title: "View Readme",
                shortcut: "ctrl+r",
                with: {repository: .full_name}
            }
        ]
    }
'
