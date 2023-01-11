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
            {type: "open-url", url: .html_url},
            {
                type: "run-script",
                script: "listPullRequests",
                title: "List Pull Requests",
                shortcut: "ctrl+p",
                with: {repository: .full_name}
            },
            {
                type: "run-script",
                script: "viewReadme",
                title: "View Readme",
                shortcut: "ctrl+r",
                with: {repository: .full_name}
            }
        ]
    }
'
