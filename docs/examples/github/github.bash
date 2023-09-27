#!/usr/bin/env bash

set -euo pipefail

# check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "gh is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
jq -n '{
    title: "GitHub",
    commands: [
        {name: "list-repos", mode: "view", title: "List Repositories"},
        {name: "list-prs", mode: "view", title: "List Pull Requests", params: [{name: "repo", type: "string"}]}
    ]
}'
exit 0
fi

COMMAND=$(jq -r .command <<< "$1")
if [ "$COMMAND" = "list-repos" ]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | jq '.[] |
        {
            title: .name,
            subtitle: (.description // ""),
            actions: [
                { title: "Open in Browser", onAction: { type: "open", url: .html_url, exit: true }},
                { title: "Copy URL", key: "o", onAction: {type: "copy",  text: .html_url, exit: true} },
                { title: "List Pull Requests", key: "p", onAction: { type: "run", command: "list-prs", params: { repo: .full_name }}}
            ]
        }
    ' | jq -s '{type: "list", items: .}'
elif [ "$COMMAND" == "list-prs" ]; then
    REPOSITORY=$(jq -r '.params.repo' <<< "$1")
    gh pr list --repo "$REPOSITORY" --json author,title,url,number | jq '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {title: "Open in Browser", onAction: { type: "open", url: .url, exit: true}},
            {title: "Copy URL", onAction: { type: "copy", text: .url, exit: true}}
        ]
    }
    ' | jq -s '{type: "list", items: .}'
fi
