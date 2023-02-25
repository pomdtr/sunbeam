#!/usr/bin/env bash

set -eo pipefail

REPO=$1

if [[ -z "$REPO" ]]; then
    echo "Usage: $0 <repo>"
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
        {type: "open-url", title: "Open in Browser", url: .url},
        {type: "copy-text", text: .url}
    ]
}
' | jq --slurp '{
    type: "list",
    items: .
}'
