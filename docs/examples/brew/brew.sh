#!/bin/bash

xargs brew search | sunbeam query -R '. | {
    title: .,
    preview: {
        command: ["brew", "info", .],
    },
    actions: [
        { type: "copy", title: "Copy", text: . },
        { type: "open", title: "Open", target: "https://formulae.brew.sh/formula/\(.)" }
    ]
}' | sunbeam query --slurp '{
    type: "list",
    generateItems: true,
    showPreview: true,
    items: (. // [])
}'
