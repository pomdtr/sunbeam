#!/usr/bin/env bash

set -eo pipefail

COMMAND="${1:-list-repos}"

if [[ $COMMAND == "list-repos" ]]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | sunbeam query --arg "command=$COMMAND" '.[] |
        {
            title: .name,
            subtitle: .description,
            accessories: [
                "\(.stargazers_count) *"
            ],
            actions: [
                {type: "open", url: .html_url},
                {
                    type: "run",
                    onSuccess: "push",
                    title: "List Pull Requests",
                    key: "p",
                    command: "\($command) list-prs \(.full_name)",
                }
            ]
        }
    ' | sunbeam query --slurp '{
        type: "list",
        items: .
    }'
elif [[ $COMMAND == "list-prs" ]]; then
    REPO=$2
    if [[ -z "$REPO" ]]; then
        echo "Usage: $0 list-prs <repo>"
        exit 1
    fi

    gh pr list --repo "$REPO" --json author,title,url,number | sunbeam query '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {type: "open", title: "Open in Browser", url: .url},
            {type: "copy", text: .url}
        ]
    }
    ' | sunbeam query --slurp '{
        type: "list",
        items: .
    }'
fi
