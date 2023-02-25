#!/bin/bash

set -euo pipefail

curl https://devdocs.io/docs/docs.json | jq '.[] |
  {
    title: .name,
    subtitle: (.release // "latest"),
    actions: [
      {
          type: "run-command",
          title: "Browse \(.release // "latest") entries",
          command: "search-entries",
          with: { slug: .slug }
      }
    ]
  }
' | jq --slurp '{ type: "list", items: . }'
