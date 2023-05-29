#!/usr/bin/env bash

DIRNAME="$(dirname "$0")"

gh gist view --filename sunbeam-extension ec9cc6b505f973f924bf025e1998cbf9 > "$DIRNAME/sunbeam-extension"
