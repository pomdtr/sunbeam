#!/bin/bash

set -euo pipefail

# check if bkt is installed
if ! command -v bkt &> /dev/null; then
  echo "bkt is not installed"
  exit 1
fi

# check if curl is installed
if ! command -v curl &> /dev/null; then
  echo "curl is not installed"
  exit 1
fi

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
  sunbeam query -n '{
    title: "DevDocs",
    root: "list-docsets",
    commands: [
      {
        name: "list-docsets",
        title: "List Docsets",
        mode: "view"
      },
      {
        name: "list-entries",
        title: "List Entries from Docset",
        mode: "view",
        params: [
          {name: "slug", type: "string", description: "docset to search"}
        ]
      }
    ]
  }'
  exit 0
fi

COMMAND=$(sunbeam query -r '.command' <<< "$1")

if [ "$COMMAND" = "list-docsets" ]; then
  # shellcheck disable=SC2016
  bkt --ttl=24h --stale=1h -- curl -s https://devdocs.io/docs/docs.json | sunbeam query '.[] |
    {
      title: .name,
      subtitle: (.release // "latest"),
      accessories: [ .slug ],
      actions: [
        {
          title: "Browse entries",
          onAction: {
            type: "run",
            command: "list-entries",
            params: {
              slug: .slug
            }
          }
        }
      ]
    }
  ' | sunbeam query -s '{ type: "list", items: .}'
elif [ "$COMMAND" = "list-entries" ]; then
  SLUG=$(sunbeam query -r '.params.slug' <<< "$1")
  # shellcheck disable=SC2016
  bkt --ttl=24h --stale=1h -- curl -s "https://devdocs.io/docs/$SLUG/index.json" | sunbeam query --arg slug="$SLUG" '.entries[] |
    {
      title: .name,
      subtitle: .type,
      actions: [
        {title: "Open in Browser", onAction: { type: "open", url: "https://devdocs.io/\($slug)/\(.path)", exit: true}},
        {title: "Copy URL", key: "c", onAction: { type: "copy", text: "https://devdocs.io/\($slug)/\(.path)", exit: true}}
      ]
    }
  ' | sunbeam query --slurp '{ type: "list", items: . }'
fi
