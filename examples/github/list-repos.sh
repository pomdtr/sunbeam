#!/usr/bin/env bash

set -eo pipefail

# shellcheck disable=SC2125
if [ -n "$1" ]; then
    ENDPOINT=/users/$1/repos?sort=updated
else
    ENDPOINT=/user/repos?sort=updated
fi

gh api "$ENDPOINT" --jq '.[] |
    {
        title: .name,
        preview: {
            command: "repo-info",
            with: {
                url: .html_url
            }
        },
        accessories: [
            "\(.stargazers_count) *"
        ],
        actions: [
            {type: "open-url", url: .html_url},
            {
                type: "run-command",
                command: "list-prs",
                title: "List Pull Requests",
                shortcut: "ctrl+p",
                with: {repository: .full_name}
            }
        ]
    }
' | sunbeam query --slurp '{
    type: "list",
    showPreview: true,
    items: .
}'
