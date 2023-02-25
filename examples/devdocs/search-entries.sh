#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2016
curl https://devdocs.io/docs/"$1"/index.json | jq --arg slug="$1" '.entries[] |
{
  title: .name,
  subtitle: .type,
  actions: [
    {type: "open-url", url: "https://devdocs.io/$slug/\(.path)"}
  ]
}
' | jq --slurp '{ type: "list", items: . }'
