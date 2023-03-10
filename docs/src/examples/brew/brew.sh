#!/bin/bash

xargs brew search | jq -R '. | {
    title: .,
    preview: {
        command: ["brew", "info", .],
    },
    actions: [
        { type: "copy", title: "Copy", text: . },
        { type: "open", title: "Open", target: "https://formulae.brew.sh/formula/\(.)" }
    ]
}' | jq --slurp '{
    type: "list",
    generateItems: true,
    showPreview: true,
    items: .
}'
