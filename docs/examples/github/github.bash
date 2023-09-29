#!/usr/bin/env bash

set -euo pipefail

# check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "gh is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
sunbeam query -n '{
    title: "GitHub",
    root: "list-repos",
    commands: [
        {name: "list-repos", mode: "view", title: "List Repositories"},
        {name: "list-prs", mode: "view", title: "List Pull Requests", params: [{name: "repo", type: "string"}]}
    ]
}'
exit 0
fi

COMMAND=$(sunbeam query -r .command <<< "$1")
if [ "$COMMAND" = "list-repos" ]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | sunbeam query '.[] |
        {
            title: .name,
            subtitle: (.description // ""),
            actions: [
                { title: "Open in Browser", onAction: { type: "open", url: .html_url, exit: true }},
                { title: "Copy URL", key: "o", onAction: {type: "copy",  text: .html_url, exit: true} },
                { title: "List Pull Requests", key: "p", onAction: { type: "run", command: "list-prs", params: { repo: .full_name }}}
            ]
        }
    ' | sunbeam query -s '{type: "list", items: .}'
elif [ "$COMMAND" == "list-prs" ]; then
    REPOSITORY=$(sunbeam query -r '.params.repo' <<< "$1")
    gh pr list --repo "$REPOSITORY" --json author,title,url,number | sunbeam query '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {title: "Open in Browser", onAction: { type: "open", url: .url, exit: true}},
            {title: "Copy URL", key: "c", onAction: { type: "copy", text: .url, exit: true}}
        ]
    }
    ' | sunbeam query -s '{type: "list", items: .}'
fi
