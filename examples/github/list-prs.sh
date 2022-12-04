#!/usr/bin/env bash

set -eo pipefail

REPO=$1

if [[ -z "$REPO" ]]; then
    echo "Usage: $0 <repo>"
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
        {type: "openUrl", title: "Open in Browser", url: .url},
        {type: "copyText", shortcut: "ctrl+y", text: .url}
    ]
}
'
