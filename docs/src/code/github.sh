#!/usr/bin/env bash

set -eo pipefail

COMMAND="${1:-list-repos}"

if [[ $COMMAND == "list-repos" ]]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | jq --arg command "$0" '.[] |
        {
            title: .name,
            subtitle: .description,
            accessories: [
                "\(.stargazers_count) *"
            ],
            actions: [
                {type: "open", target: .html_url},
                {
                    type: "push",
                    title: "List Pull Requests",
                    shortcut: "ctrl+p",
                    command:[$command, "list-prs", .full_name]
                }
            ]
        }
    ' | jq --slurp '{
        type: "list",
        actions: [
            {type: "open", title: "Open Github", target: "https://github.com"}
        ],
        items: .
    }'
elif [[ $COMMAND == "list-prs" ]]; then
    REPO=$2
    if [[ -z "$REPO" ]]; then
        echo "Usage: $0 list-prs <repo>"
        exit 1
    fi

    gh pr list --repo "$REPO" --json author,title,url,number | jq '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {type: "open", title: "Open in Browser", target: .url},
            {type: "copy", text: .url}
        ]
    }
    ' | jq --slurp '{
        type: "list",
        items: .
    }'
fi
