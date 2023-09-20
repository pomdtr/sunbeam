#!/usr/bin/env bash

set -euo pipefail

# check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
jq -n '{
    title: "GitHub",
    commands: [
        {output: "list", name: "list-repos", title: "List Repositories"},
        {output: "list", name: "list-prs", title: "List Pull Requests", params: [{name: "repository", type: "string"}]}
    ]
}'
exit 0
fi

if [ "$1" = "list-repos" ]; then
    # shellcheck disable=SC2016
    gh api "/user/repos?sort=updated" | jq '.[] |
        {
            title: .name,
            subtitle: (.description // ""),
            actions: [
                { type: "open", title: "Open in Browser", url: .html_url },
                { type: "copy", title: "Copy URL", text: .html_url, key: "o" },
                { type: "run", title: "List Pull Requests", key: "p", command: { name: "list-prs", params: { repository: .full_name }}}
            ]
        }
    ' | jq -s '{items: .}'
elif [ "$1" == "list-prs" ]; then
    eval "$(sunbeam parse bash)"

    gh pr list --repo "$REPOSITORY" --json author,title,url,number | jq '.[] |
    {
        title: .title,
        subtitle: .author.login,
        accessories: [
            "#\(.number)"
        ],
        actions: [
            {type: "open", title: "Open in Browser", url: .url},
            {type: "copy", title: "Copy URL", text: .url}
        ]
    }
    ' | jq -s '{items: .}'
fi
