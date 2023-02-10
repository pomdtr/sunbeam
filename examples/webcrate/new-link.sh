#!/bin/bash
set -x
set -eo pipefail

LINK_URL="$1"

curl \
    -X POST \
    -H X-Space-App-Key:"$WEBCRATE_API_KEY" \
    -H Content-Type:application/json \
    -d '{"url": "'"$LINK_URL"'"}' \
    "$WEBCRATE_URL"/api/link

