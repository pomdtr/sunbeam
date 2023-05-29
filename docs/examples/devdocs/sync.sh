#!/usr/bin/env bash

DIRNAME="$(dirname "$0")"

gh gist view --filename sunbeam-extension 287f173468d1d0e43a43972729d513ec > "$DIRNAME/sunbeam-extension"
