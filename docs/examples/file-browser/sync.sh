#!/usr/bin/env bash

DIRNAME="$(dirname "$0")"

gh gist view --filename sunbeam-extension 59cac008e26986dcfe9e8661d084bca5 > "$DIRNAME/sunbeam-extension"
