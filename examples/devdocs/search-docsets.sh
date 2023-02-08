#!/bin/bash

set -euo pipefail

sunbeam http https://devdocs.io/docs/docs.json | sunbeam query '.[] |
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
' | sunbeam query --slurp '{ type: "list", items: . }'
