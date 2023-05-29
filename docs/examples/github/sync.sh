#!/usr/bin/env bash

DIRNAME="$(dirname "$0")"

gh gist view --filename sunbeam-extension dc1363d8f641928893ca8d3e670c9c3d > "$DIRNAME/sunbeam-extension"
