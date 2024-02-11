#!/bin/sh

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Brew",
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
                        type: "string"
                    }
                ]
            }

        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "list" ]; then
    brew list | jq -R '{
        title: .,
        actions: [
            {
                title: "Uninstall",
                command: "uninstall",
                reload: true,
                params: {
                    package: .
                }
            }
        ]
    }' | jq -s '{ items: . }'
elif [ "$COMMAND" = "uninstall" ]; then
    PACKAGE=$(echo "$1" | jq -r '.params.package')
    brew uninstall "$PACKAGE"
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
