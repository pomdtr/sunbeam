#!/usr/bin/env bash

# check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq could not be found"
    exit
fi

# check is sqlite3 is installed
if ! command -v sqlite3 &> /dev/null
then
    echo "sqlite3 could not be found"
    exit
fi

if [ $# -eq 0 ]
then
    jq -n '{
        title: "VS Code",
        commands: [
            {name: "list-projects", title: "List Projects", mode: "page"},
            {name: "open-project", title: "Open Project", mode: "silent", params: [{name: "dir", type: "string"}]}
        ]
    }'
    exit 0
fi

INPUT="$(cat)"
if [ "$1" = "list-projects" ]; then
    dbPath="$HOME/Library/Application Support/Code/User/globalStorage/state.vscdb"
    query="SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'"

    # get the recently opened paths
    sqlite3 "$dbPath" "$query" | jq '.[] | select(.folderUri) | {
        title: (.folderUri | split("/") | last),
        actions: [
            {type: "run", title: "Open in VS Code", command: {name: "open-project", params: {dir: (.folderUri | sub("^file://"; ""))}}},
            {type: "open", title: "Open Folder", url: .folderUri}

        ]
    }' | jq -s '{
        type: "list",
        items: .
    }'
elif [ "$1" = "open-project" ]; then
    dir="$(echo "$INPUT" | jq -r '.params.dir')"
    code "$dir"
fi


