#!/usr/bin/env bash

DIRNAME=$(dirname "$0")

sunbeam query -R --slurp '{
    type: "detail",
    title: "Highlight",
    preview: {
        text: .,
        language: "markdown"
    }
}' < "$DIRNAME/highlight.md"
