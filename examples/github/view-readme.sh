#!/bin/bash

set -euo pipefail

REPO=$1

# shellcheck disable=SC2016
gh api "repos/$REPO/readme" | sunbeam query --arg REPO="$REPO" '
{
  type: "detail",
  preview: {
    command: "repo-info",
    with: {
      url: "https://github.com/\($REPO)"
    }
  },
  actions: [
    { type: "open-url", title: "Open in Browser", url: .html_url }
  ]
  }
'
