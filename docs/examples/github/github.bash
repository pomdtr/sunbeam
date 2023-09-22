#!/usr/bin/env bash

set -euo pipefail

# check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

# check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "gh is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
jq -n '{
    title: "GitHub",
    commands: [
        {name: "list-repos", mode: "page", title: "List Repositories"},
        {name: "list-prs", mode: "page", title: "List Pull Requests", params: [{name: "repo", type: "string"}]}
    ]
}'
exit 0
fi

INPUT=$(cat)
if [ "$1" = "list-repos" ]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | jq '.[] |
        {
            title: .name,
            subtitle: (.description // ""),
            actions: [
                { type: "open", title: "Open in Browser", url: .html_url, exit: true },
                { type: "copy", title: "Copy URL", text: .html_url, exit: true, key: "o" },
                { type: "run", title: "List Pull Requests", key: "p", command: { name: "list-prs", params: { repo: .full_name }}}
            ]
        }
    ' | jq -s '{type: "list", items: .}'
elif [ "$1" == "list-prs" ]; then
    REPOSITORY=$(jq -r '.params.repo' <<< "$INPUT")
    gh pr list --repo "$REPOSITORY" --json author,title,url,number | jq '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {type: "open", title: "Open in Browser", url: .url, exit: true},
            {type: "copy", title: "Copy URL", text: .url, exit: true}
        ]
    }
    ' | jq -s '{type: "list", items: .}'
fi
