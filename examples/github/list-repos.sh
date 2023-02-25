#!/usr/bin/env bash

set -eo pipefail

# shellcheck disable=SC2125
if [ -n "$1" ]; then
    ENDPOINT=/users/$1/repos?sort=updated
else
    ENDPOINT=/user/repos?sort=updated
fi

# shellcheck disable=SC2016
gh api "$ENDPOINT" --jq '.[] |
    {
        title: .name,
        subtitle: .description,
        accessories: [
            "\(.stargazers_count) *"
        ],
        actions: [
            {type: "open-url", url: .html_url},
            {
                type: "run-command",
                command: "view-readme",
                title: "View README",
                with: {repository: .full_name}
            },
            {
                type: "run-command",
                command: "list-prs",
                title: "List Pull Requests",
                with: {repository: .full_name}
            }
        ]
    }
' | jq --arg "repo=$1" --slurp '{
    type: "list",
    title: "List \($repo) Repositories",
    items: .
}'
