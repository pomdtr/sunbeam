#!/usr/bin/env -S sunbeam shell

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Extensions",
        root: [
            { command: "list" },
            { command: "create" }
        ],
        commands: [
            {
                name: "list",
                title: "Manage Extensions",
                mode: "list",
            },
            {
                name: "delete",
                title: "Delete Extensions",
                mode: "silent",
                params: [
                    { name: "alias", title: "Alias", type: "text", required: true }
                ]
            },
            {
                name: "install",
                title: "Install Extension",
                mode: "silent",
                params: [
                    { name: "alias", title: "Alias", type: "text", required: true },
                    { name: "url", title: "URL", type: "text", required: true }
                ]
            }
        ]
    }'
    exit 0
fi

if [ -n "$SUNBEAM_CONFIG" ]; then
    CONFIG_PATH="$SUNBEAM_CONFIG"
elif [ -n "$XDG_CONFIG_HOME" ]; then
    CONFIG_PATH="$XDG_CONFIG_HOME/sunbeam/config.json"
else
    CONFIG_PATH="$HOME/.config/sunbeam/config.json"
fi


COMMAND="$(echo "$1" | sunbeam query -r ".command")"
if [ "$COMMAND" = "list" ]; then
    sunbeam query '.extensions | to_entries | {
        items: map({
            title: .key,
            accessories: [.value],
            actions: [
                { title: "Copy URL", type: "copy", text: .value, exit: true },
                { title: "Delete Extension", key: "d", type: "run", command: "delete", params: { alias: .key }, reload: true },
                { title: "Install Extension", key: "n", type: "run", command: "install", reload: true }
            ]
        })
    }' "$CONFIG_PATH"
elif [ "$COMMAND" = "delete" ]; then
    ALIAS=$(echo "$1" | sunbeam query -r ".params.alias")
    # shellcheck disable=SC2016
    sunbeam query --in-place --arg alias="$ALIAS" 'del(
        .extensions[$alias]
    )' "$CONFIG_PATH"
elif [ "$COMMAND" = "create" ]; then
    ALIAS=$(echo "$1" | sunbeam query -r ".params.alias")
    URL=$(echo "$1" | sunbeam query -r ".params.url")

    # shellcheck disable=SC2016
    sunbeam query --in-place --arg url="$URL" --arg alias="$ALIAS" '.extensions += { $alias : $url }' "$CONFIG_PATH"
fi
