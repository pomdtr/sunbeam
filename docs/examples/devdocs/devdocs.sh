#!/bin/bash

set -euo pipefail

if [ $# -eq 1 ]; then
  # shellcheck disable=SC2016
  curl "https://devdocs.io/docs/$1/index.json" | sunbeam query --arg slug="$1" '.entries[] |
{
  title: .name,
  subtitle: .type,
  actions: [
    {type: "open", url: "https://devdocs.io/\($slug)/\(.path)"}
  ]
}
' | sunbeam query --slurp '{ type: "list", items: . }'

  exit 0
fi

# shellcheck disable=SC2016
curl https://devdocs.io/docs/docs.json | sunbeam query --arg command="$0" '.[] |
  {
    title: .name,
    subtitle: (.release // "latest"),
    accessories: [ .slug ],
    actions: [
      {
          type: "run",
          "onSuccess": "push",
          title: "Browse \(.release // "latest") entries",
          command: "\($command) \(.slug)",
      }
    ]
  }
' | sunbeam query --slurp '{ type: "list", items: . }'
