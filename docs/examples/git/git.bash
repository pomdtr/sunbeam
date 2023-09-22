#!/usr/bin/env bash

# check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed"
    exit 1
fi

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
    jq -n '{
        title: "Git",
        commands: [
            {
                name: "list-commits",
                title: "List Commits",
                mode: "page"
            },
            {
                name: "commit-form",
                title: "Write Commit",
                mode: "page"
            },
            {
                name: "commit",
                title: "Commit",
                mode: "silent",
                params: [
                    {name: "message", type: "string", description: "Commit Message"}
                ]
            }
        ]
    }'
    exit 0
fi

INPUT=$(cat)
COMMAND=$1

if [ "$COMMAND" = "list-commits" ]; then
    git log | jc --git-log | jq '.[] | {
        title: .message,
        subtitle: .author,
        accessories: [
            .commit[:7]
        ],
    }' | jq -s '{type: "list", items: .}'
elif [ "$COMMAND" = "commit-form" ]; then
    jq -n '{
        type: "form",
        command: {"name": "commit"},
        inputs: [
            {name: "message", type: "textarea", title: "Commit Message"}
        ]
    }'
elif [ "$COMMAND" = "commit" ]; then
    MESSAGE=$(jq -r '.params.message' <<< "$INPUT")
    git commit -m "$MESSAGE"
fi

