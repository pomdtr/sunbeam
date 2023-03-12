#!/usr/bin/env bash

# shellcheck disable=SC2016
sunbeam query -n '{
    type: "list",
    title: "Highlight",
    showDetail: true,
    items: [
        {
            title: "Highlight",
            detail: {
                text: "**test**, *test*, `test`",
                language: "markdown"
            }
        }
    ]
}'
