#!/bin/sh

set -eu

# view source at https://val.town/v/pomdtr/sunbeam_example
REMOTE_URL="https://pomdtr-sunbeam_example.web.val.run/"

# check if curl is installed
if ! [ -x "$(command -v curl)" ]; then
    echo "curl is not installed. Please install it." >&2
    exit 1
fi

# if no args
if [ $# -eq 0 ]; then
    curl -s "$REMOTE_URL"
    exit 0
fi

exec curl -X POST --data "@-" "$REMOTE_URL$1"
