#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Val Town",
        description: "Manage your Vals",
        commands: [
            {
                name: "home",
                title: "List Home Vals",
                mode: "list"
            },
            {
                name: "create",
                title: "Create Val",
                mode: "tty",
            },
            {
                name: "edit",
                title: "Edit Val",
                mode: "tty",
                params: [
                    {
                        name: "val",
                        type: "string",
                        title: "Val ID",
                        required: true
                    }
                ]
            }
        ]
    }'
    exit 0
fi

if [ -z "$VALTOWN_TOKEN" ]; then
    echo "VALTOWN_TOKEN is not set"
    exit 1
fi

API_ROOT="https://api.val.town"
COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "home" ]; then
    USER_ID=$(sunbeam fetch -H "Authorization: Bearer $VALTOWN_TOKEN" "$API_ROOT/v1/me" | sunbeam query -r '.id')
    sunbeam fetch -H "Autorization: Bearer $VALTOWN_TOKEN" "$API_ROOT/v1/users/$USER_ID/vals" | sunbeam query '.data | {

        items: map({
            title: .name,
            subtitle: "v\(.version)",
            accessories: [.privacy],
            actions: [
                {
                    title: "Open in Browser",
                    type: "open",
                    target: "https://val.town/v/\(.author.username[1:])/\(.name)",
                    exit: true
                },
                {
                    title: "Edit Val",
                    type: "run",
                    key: "e",
                    command: "edit",
                    params: {
                        val: .id
                    },
                    reload: true
                },
                {
                    title: "Create New Val",
                    type: "run",
                    key: "n",
                    command: "Create Val",
                    reload: true
                },
                {
                    title: "Copy URL",
                    key: "c",
                    exit: true,
                    type: "copy",
                    text: "https://val.town/v/\(.author.username[1:])/\(.name)"
                },
                {
                    title: "Copy Web Endpoint",
                    exit: true,
                    key: "w",
                    type: "copy",
                    text: "https://\(.author.username[1:])-\(.name).web.val.run"
                },
                {
                    "title": "Copy Run Endpoint",
                    "key": "r",
                    exit: true,
                    type: "copy",
                    text: "https://api.val.town/v1/run/\(.author.username[1:])/\(.name)"
                }
            ]
        })
    }'
elif [ "$COMMAND" = "create" ]; then
    printf "export function untitled() {}" | sunbeam edit -e tsx \
        | sunbeam query -Rs '{code: .}' \
        | sunbeam fetch \
            -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $VALTOWN_TOKEN" \
            -d @- \
            "$API_ROOT/v1/vals" > /dev/null
elif [ "$COMMAND" = "edit" ]; then
    VAL_ID=$(echo "$1" | sunbeam query -r '.params.val')

    sunbeam fetch -H "Authorization: Bearer $VALTOWN_TOKEN" "$API_ROOT/v1/vals/$VAL_ID" \
        | sunbeam query -r .code \
        | sunbeam edit -e tsx \
        | sunbeam query -Rs '{ code: . }' \
        | sunbeam fetch -X POST -d @- -H "Content-Type: application/json" -H "Authorization: Bearer $VALTOWN_TOKEN" "$API_ROOT/v1/vals/$VAL_ID/versions"
fi

