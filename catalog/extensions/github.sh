#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
sunbeam query -n '{
    title: "GitHub",
    description: "Search GitHub Repositories",
    commands: [
        {name: "search-repos", mode: "list", title: "Search Repositories"},
        {name: "list-prs", mode: "list", title: "List Pull Requests", params: [{name: "repo", type: "string", required: true}]}
    ]
}'
exit 0
fi

# check if gh is installed
if ! command -v gh >/dev/null 2>&1; then
    echo "gh is not installed"
    exit 1
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "search-repos" ]; then
    # shellcheck disable=SC2016
    QUERY=$(echo "$1" | sunbeam query '.query')
    if [ "$QUERY" = "null" ]; then
        gh api "/user/repos?sort=updated" | sunbeam query '{
            dynamic: true,
            items: map({
                title: .full_name,
                subtitle: (.description // ""),
                actions: [
                    { title: "Open in Browser", type: "open", target: .html_url, exit: true },
                    { title: "Copy URL", key: "o", type: "copy",  text: .html_url, exit: true},
                    { title: "List Pull Requests", key: "p", type: "run", command: "list-prs", params: { repo: .full_name }}
                ]
            })
        }'
        exit 0
    fi
    gh api "search/repositories?q=$QUERY" | sunbeam query '.items | {
        dynamic : true,
        items: map({
                title: .full_name,
                subtitle: (.description // ""),
                actions: [
                    { title: "Open in Browser", type: "open", target: .html_url, exit: true },
                    { title: "Copy URL", key: "o", type: "copy",  text: .html_url, exit: true},
                    { title: "List Pull Requests", key: "p", type: "run", command: "list-prs", params: { repo: .full_name }}
                ]
            })
        }'
elif [ "$1" = "list-prs" ]; then
    REPOSITORY=$(echo "$1" | sunbeam query -r '.params.repo')
    gh pr list --repo "$REPOSITORY" --json author,title,url,number | sunbeam query ' {
        items: map({
            title: .title,
            subtitle: .author.login,
            accessories: [
                "#\(.number)"
            ],
            actions: [
                {title: "Open in Browser", type: "open", target: .url, exit: true},
                {title: "Copy URL", key: "c", type: "copy", text: .url, exit: true}
            ]
        })
    }'
fi
