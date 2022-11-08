#!/usr/bin/env bash

set -eo pipefail

OWNER=$1

gh repo list "$OWNER" --json nameWithOwner,url,stargazerCount,owner | sunbeam jq '.[] |
    {
        title: .nameWithOwner,
        subtitle: .owner.login,
        accessories: [
            "\(.stargazerCount) *"
        ],
        actions: [
            {type: "open-url", url: .url},
            {
                type: "push-page",
                title: "List Pull Requests",
                shortcut: "ctlr+p",
                page: "list-pull-requests",
                with: {repository: .nameWithOwner}
            }
        ]
    }
'
