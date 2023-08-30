#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2016
tldr --list | sunbeam query -R '{
    title: .,
    actions: [
      {
        type: "push",
        command: {
          name: "view",
          args: {
            command: "\(.)"
          }
        }
      },
      {
        type: "open",
        title: "Open in browser",
        target: "https://tldr.inbrowser.app/pages/common/\(. | @uri)",
      }
    ]
}'
