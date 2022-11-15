#!/usr/bin/env bash

set -eo pipefail

if [ -n "$1" ]; then
    ENDPOINT=/users/$1/repos
else
    ENDPOINT=/user/repos
fi

gh api "$ENDPOINT" --paginate --cache 3h --jq '.[] |
    {
        title: .full_name,
        subtitle: .owner.login,
        accessories: [
            "\(.stargazers_count) ⭐"
        ],
        actions: [
            {type: "open-url", url: .html_url},
            {
                type: "push-page",
                title: "List Pull Requests",
                shortcut: "ctlr+p",
                script: "list-pull-requests",
                with: {repository: .full_name}
            }
        ]
    }
'
