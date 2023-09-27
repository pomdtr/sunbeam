#!/usr/bin/env bash

sunbeam fetch "https://www.nixhub.io/search?q=$1&_data=routes/_nixhub.search" \
    | jq '.results[] | {
        title: .name,
        subtitle: .summary,
        actions: [{type: "open", title: "Open in Browser", target: "https://www.nixhub.io/packages/\(.name)" }]
      }' \
    | sunbeam list --json --title "Search Nix Packages"
