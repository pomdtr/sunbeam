#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Meteo",
        description: "Show Meteo",
        commands: [
            {
                name: "show",
                title: "Show Meteo",
                mode: "page"
            }
        ]
    }'
    exit 0
fi

sunbeam fetch -A "curl/7.75.0 (x86_64-pc-linux-gnu)" 'wttr.in/?3n' | sunbeam query -sR '{
    type: "detail",
    text: .
}'
