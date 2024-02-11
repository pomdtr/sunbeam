#!/bin/sh

set -e

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
  jq -n '{
    title: "DevDocs",
    description: "Search DevDocs.io",
    commands: [
      {
        name: "list-docsets",
        title: "List Docsets",
        mode: "filter"
      },
      {
        name: "list-entries",
        title: "List Entries from Docset",
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
        {title: "Open in Browser", extension: "std", command: "open", params: {url: "https://devdocs.io/\($slug)/\(.path)"}},
        {title: "Copy URL", extension: "std", command: "copy", params: {text: "https://devdocs.io/\($slug)/\(.path)"}}
      ]
    }) | {  items: . }'
fi
