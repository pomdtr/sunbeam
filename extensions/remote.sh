#!/bin/sh

set -eu

# view source at https://val.town/v/pomdtr/sunbeam_example
REMOTE_URL="https://pomdtr-sunbeam_example-cync-9.web.val.run/"

# check if curl is installed
if ! [ -x "$(command -v curl)" ]; then
    echo "curl is not installed. Please install it." >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    exec curl -s "$REMOTE_URL"
fi

exec curl -s -X POST -d "@-" "$REMOTE_URL$1"
