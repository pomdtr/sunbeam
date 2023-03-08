#!/bin/bash

set -x

set -euo pipefail

if [ $# -eq 1 ]; then
  curl "https://devdocs.io/docs/$1/index.json" | jq --arg slug "$1" '.entries[] |
{
  title: .name,
  subtitle: .type,
  actions: [
    {type: "open", target: "https://devdocs.io/\($slug)/\(.path)"}
  ]
}
' | jq --slurp '{ type: "list", items: . }'

  exit 0
fi

curl https://devdocs.io/docs/docs.json | jq --arg command "$0" '.[] |
  {
    title: .name,
    subtitle: (.release // "latest"),
    accessories: [ .slug ],
    actions: [
      {
          type: "push",
          title: "Browse \(.release // "latest") entries",
          command: $command,
          args: [ .slug ],
      }
    ]
  }
' | jq --slurp '{ type: "list", items: . }'
