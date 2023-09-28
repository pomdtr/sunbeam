#!/usr/bin/env bash

# check if jc is installed
if ! command -v jc &> /dev/null; then
    echo "jc is not installed"
    exit 1
fi

# check if git is installed
if ! command -v git &> /dev/null; then
    echo "git is not installed"
    exit 1
fi

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Git",
        commands: [
            {
                name: "list-commits",
                title: "List Commits",
                mode: "view"
            },
            {
                name: "commit",
                title: "Write Commit",
                mode: "view"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(sunbeam query -r '.command' <<< "$1")
if [ "$COMMAND" = "list-commits" ]; then
    git log | jc --git-log | sunbeam query '.[] | {
        title: .message,
        subtitle: .author,
        accessories: [
            .commit[:7]
        ],
        actions: [
            {  title: "Copy Commit Hash", onAction: { type: "copy", text: .commit, exit: true }},
            {
                title: "New Commit", key: "n", onAction: {
                    type: "run",
                    command: "commit"
                }
            }
        ]
    }' | sunbeam query -s '{type: "list", items: .}'
elif [ "$COMMAND" = "commit" ]; then
    INPUTS=$(sunbeam query .inputs <<< "$1")
    if [ "$INPUTS" = "null" ]; then
        sunbeam query -n '{
            type: "form",
            items: [
                {name: "message", type: "textarea", title: "Commit Message"}
            ]
        }'
        exit 0
    fi

    MESSAGE=$(sunbeam query -r '.inputs.message' <<< "$1")
    echo "$MESSAGE" > .commit-message
fi
