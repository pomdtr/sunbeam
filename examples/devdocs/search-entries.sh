#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2016
sunbeam http https://devdocs.io/docs/"$1"/index.json | sunbeam query --arg slug="$1" '.entries[] |
{
  title: .name,
  subtitle: .type,
  actions: [
    {type: "open-url", url: "https://devdocs.io/$slug/\(.path)"}
  ]
}
' | sunbeam query --slurp '{ type: "list", items: . }'
