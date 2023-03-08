#!/bin/bash

set -euo pipefail

PLATFORM="${1:-macos}"

# shellcheck disable=SC2016
tldr --list --platform="$PLATFORM" | jq --arg platform "$PLATFORM" -R '{
    title: .,
    preview: {
      command: "tldr",
      args: ["--color",  "always",  "--platform", $platform, .]
    }
}' | jq --slurp '
  {
    type: "list",
    showPreview: true,
    items: .
  }
'
