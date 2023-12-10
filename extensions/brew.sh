#!/bin/sh

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Brew",
        root: ["list"],
        commands: [
            {
                name: "list",
                title: "List Installed Packages",
                mode: "filter",
            },
            {
                name: "uninstall",
                title: "Uninstall Package",
                mode: "silent",
                params: [
                    {
                        name: "package",
                        title: "Package Name",
                        type: "text"
                    }
                ]
            }

        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | sunbeam query -r '.command')
if [ "$COMMAND" = "list" ]; then
    brew list | sunbeam query -R '{
        title: .,
        actions: [
            {
                title: "Uninstall",
                type: "run",
                command: "uninstall",
                params: {
                    package: .
                },
                reload: true
            }
        ]
    }' | sunbeam query -s '{ items: . }'
elif [ "$COMMAND" = "uninstall" ]; then
    PACKAGE=$(echo "$1" | sunbeam query -r '.params.package')
    brew uninstall "$PACKAGE"
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
