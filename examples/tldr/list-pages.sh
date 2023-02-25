#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2016
tldr --list --platform="$1" | jq --arg platform="$1" -R '{
    title: .,
    preview: {
      command: "view-page",
      with: {page: ., platform: $platform}
    },
    actions: [
      {type: "run-command", "command": "view-page", title: "View Page", with: {page: ., platform: $platform}}
    ]
}' | jq --slurp '
  {
    type: "list",
    actions: [
      {
        title: "Refresh Pages",
        type: "run-command",
        command: "update",
        onSuccess: "reload-page"
      }
    ],
    showPreview: true,
    items: .
  }
'
