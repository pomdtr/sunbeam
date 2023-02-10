#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2016
tldr --list --platform="$1" | sunbeam query --arg platform="$1" -R '{
    title: .,
    preview: {
      command: "view-page",
      with: {page: ., platform: $platform}
    },
    actions: [
      {type: "run-command", "command": "view-page", title: "View Page", with: {page: ., platform: $platform}}
    ]
}' | sunbeam query --slurp '
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
