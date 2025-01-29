#!/bin/sh

if ! [ -x "$(command -v uv)" ]; then
    echo "uv is not installed. Please install it." >&2
    exit 1
fi

exec uv run https://raw.githubusercontent.com/pomdtr/sunbeam/10ace61/extensions/files.py "$@"
