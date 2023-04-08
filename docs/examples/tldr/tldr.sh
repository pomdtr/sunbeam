#!/bin/bash

set -euo pipefail

PLATFORM="${1:-osx}"

# shellcheck disable=SC2016
tldr --list --platform="$PLATFORM" | sunbeam query --arg platform="$PLATFORM" -R '{
    title: .,
    preview: {
      command: "tldr --color=always --platform=\($platform) \(.)",
    },
    actions: [
      {
        type: "open-url",
        title: "Open in browser",
        url: "https://tldr.inbrowser.app/pages/common/\(. | @uri)",
      }
    ],
}' | sunbeam query --slurp '
  {
    type: "list",
    showPreview: true,
    items: .
  }
'
