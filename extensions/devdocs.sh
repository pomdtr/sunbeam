#!/bin/sh

set -e

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
  jq -n '{
    title: "DevDocs",
    description: "Search DevDocs.io",
    root: [
      { title: "Search Docsets", type: "run", command: "list-docsets" },
      { title: "Search Golang Entries", type: "run", command: "list-entries", params: { slug: "go" } }
    ],
    commands: [
      {
        name: "list-docsets",
        mode: "filter"
      },
      {
        name: "list-entries",
        mode: "filter",
        params: [
          { name: "slug", title: "Slug", type: "string" }
        ]
      }
    ]
  }'
  exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list-docsets" ]; then
  # shellcheck disable=SC2016
  curl https://devdocs.io/docs/docs.json | jq 'map({
      title: .name,
      subtitle: (.release // "latest"),
      accessories: [ .slug ],
      actions: [
        {
          title: "Browse entries",
          type: "run",
          command: "list-entries",
          params: {
            slug: .slug
          }
        }
      ]
    }) | {  items: . }'
elif [ "$COMMAND" = "list-entries" ]; then
  SLUG=$(echo "$1" | jq -r '.params.slug')
  # shellcheck disable=SC2016
  curl "https://devdocs.io/docs/$SLUG/index.json" | jq --arg slug "$SLUG" '.entries | map({
      title: .name,
      subtitle: .type,
      actions: [
        {title: "Open in Browser", type: "open", target: "https://devdocs.io/\($slug)/\(.path)"},
        {title: "Copy URL", key: "c", type: "copy", text: "https://devdocs.io/\($slug)/\(.path)"}
      ]
    }) | {  items: . }'
fi
