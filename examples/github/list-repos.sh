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
        title: .full_name,
        subtitle: .owner.login,
        preview: (.description // "No description"),
        accessories: [
            "\(.stargazers_count) *"
        ],
        actions: [
            {type: "openUrl", url: .html_url},
            {
                type: "runScript",
                script: "listPullRequests",
                title: "List Pull Requests",
                shortcut: "ctrl+p",
                with: {repository: .full_name}
            },
            {
                type: "runScript",
                script: "viewReadme",
                title: "View Readme",
                shortcut: "ctrl+r",
                with: {repository: .full_name}
            }
        ]
    }
'
