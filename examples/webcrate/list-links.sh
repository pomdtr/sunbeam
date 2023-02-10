#!/bin/bash
set -x
set -eo pipefail

curl \
    -H X-Space-App-Key:"$WEBCRATE_API_KEY" \
    "$WEBCRATE_URL/api/link" | sunbeam query '.data[] | {
        title: .meta.title,
        subtitle: .meta.description,
        actions: [
            { title: "Open in Browser", type: "open-url", url: .url },
            { title: "Copy URL", type: "copy-text", text: .url }
        ]
    }' | sunbeam query --slurp '{
        type: "list",
        items: .,
        actions: [
            { title: "New Link", type: "run-command", command: "new-link"}
        ]
    }'


