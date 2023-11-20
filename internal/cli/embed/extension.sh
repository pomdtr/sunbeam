#!/bin/sh

if [ "$#" -eq 0 ]; then
    sunbeam query -n '{
        title: "My Extension",
        description: "This is my extension",
        items: [
            {
                title: "Hi Mom!",
                command: "hi",
                params: {
                    name: "World"
                }
            },
            {
                command: "hi",
            }
        ],
        commands: [
            {
                name: "hi",
                title: "Say Hi",
                mode: "detail",
                params: [
                    {
                        name: "name",
                        title: "Name",
                        type: "text",
                        required: true
                    }
                ]
            }
        ]
    }'
    exit 0
fi

payload="$1"

COMMAND=$(echo "$payload" | sunbeam query -r '.command')
if [ "$COMMAND" = "hi" ]; then
    name="$(echo "$payload" | sunbeam query -r '.params.name')"
    # shellcheck disable=SC2016
    sunbeam query -n --arg name "$name" '{
        text: "Hi \($name)!",
        actions: [
            {
                title: "Copy Name",
                type: "copy",
                text: $name
            }
        ]
    }'
else
    echo "Unknown command: $COMMAND" >&2
    exit 1
fi
