#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Cli Apps",
        description: "Open your favorite cli apps",
        commands: [
            {
                name: "htop",
                title: "htop",
                mode: "tty"
            },
            {
                name: "btop",
                title: "btop",
                mode: "tty"
            },
            {
                name: "cmatrix",
                title: "cmatrix",
                mode: "tty"
            },
            {
                name: "cowsay",
                title: "cowsay",
                mode: "tty"
            },
            {
                name: "asciiquarium",
                title: "asciiquarium",
                mode: "tty"
            },
            {
                name: "lf",
                title: "lf",
                mode: "tty"
            },
            {
                name: "aichat",
                title: "aichat",
                mode: "tty"
            },
            {
                name: "node",
                title: "Open Node Shell",
                mode: "tty"
            },
            {
                name: "fish",
                title: "Open Fish Shell",
                mode: "tty"
            },
            {
                name: "python",
                title: "Open Python Shell",
                mode: "tty"
            },
            {
                name: "sncli",
                title: "Open sncli",
                mode: "tty"
            },
            {
                name: "yaegi",
                title: "Open Yaegi",
                mode: "tty"
            },
            {
                name: "http-prompt",
                title: "Open http-prompt",
                mode: "tty"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "htop" ]; then
    htop
elif [ "$COMMAND" = "btop" ]; then
    btop
elif [ "$COMMAND" = "cmatrix" ]; then
    cmatrix
elif [ "$COMMAND" = "cowsay" ]; then
    cowsay "Sunbeam is awesome!" | less
elif [ "$COMMAND" = "asciiquarium" ]; then
    asciiquarium
elif [ "$COMMAND" = "lf" ]; then
    lf
elif [ "$COMMAND" = "aichat" ]; then
    aichat
elif [ "$COMMAND" = "node" ]; then
    node
elif [ "$COMMAND" = "python" ]; then
    ipython3
elif [ "$COMMAND" = "fish" ]; then
    fish
elif [ "$COMMAND" = "sncli" ]; then
    sncli
elif [ "$COMMAND" = "yaegi" ]; then
    rlwrap yaegi || true
elif [ "$COMMAND" = "butterfish" ]; then
    butterfish shell
elif [ "$COMMAND" = "http-prompt" ]; then
    http-prompt
fi
