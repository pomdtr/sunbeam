#!/usr/bin/env bash

DIRNAME="$(dirname "$0")"

gh gist view --filename sunbeam-extension bf9781444318cd9a5845444a6ac4f467 > "$DIRNAME/sunbeam-extension"
