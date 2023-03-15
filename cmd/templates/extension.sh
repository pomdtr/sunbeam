#!/usr/bin/env bash

set -eo pipefail

sunbeam query --null-input '{
    type: "detail",
    title: "Hello, world!",
    text: "Hello, world!",
    actions: [
        {
            type: "copy",
            text: "Hello, world!",
        }
    ]
}'
