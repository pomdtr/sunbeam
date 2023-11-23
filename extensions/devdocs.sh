#!/bin/sh

set -e

# if no arguments are passed, return the extension's manifest
if [ $# -eq 0 ]; then
  sunbeam query -n '{
    title: "DevDocs",
    description: "Search DevDocs.io",
    root: [ "list-docsets" ],
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
          { name: "slug", title: "Slug", type: "text", required: true }
        ]
      }
    ]
  }'
  exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "list-docsets" ]; then
  # shellcheck disable=SC2016
  sunbeam fetch https://devdocs.io/docs/docs.json | sunbeam query 'map({
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
  SLUG=$(echo "$1" | sunbeam query -r '.params.slug')
  # shellcheck disable=SC2016
  sunbeam fetch "https://devdocs.io/docs/$SLUG/index.json" | sunbeam query --arg slug="$SLUG" '.entries | map({
      title: .name,
      subtitle: .type,
      actions: [
        {title: "Open in Browser", type: "open", target: "https://devdocs.io/\($slug)/\(.path)", exit: true},
        {title: "Copy URL", key: "c", type: "copy", text: "https://devdocs.io/\($slug)/\(.path)", exit: true}
      ]
    }) | {  items: . }'
fi
