#!/usr/bin/env bash

# check is sqlite3 is installed
if ! command -v sqlite3 &> /dev/null
then
    echo "sqlite3 could not be found"
    exit
fi

if [ $# -eq 0 ]
then
    sunbeam query -n '{
        title: "VS Code",
        root: "list-projects",
        commands: [
            {name: "list-projects", title: "List Projects", mode: "view"},
            {name: "open-project", title: "Open Project", mode: "no-view", params: [{name: "dir", type: "string"}]}
        ]
    }'
    exit 0
fi

COMMAND=$(sunbeam query -r .command <<< "$1")
if [ "$COMMAND" = "list-projects" ]; then
    dbPath="$HOME/Library/Application Support/Code/User/globalStorage/state.vscdb"
    query="SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'"

    # get the recently opened paths
    sqlite3 "$dbPath" "$query" | sunbeam query '.[] | select(.folderUri) | {
        title: (.folderUri | split("/") | last),
        actions: [
            {title: "Open in VS Code", onAction: { type: "run", command: "open-project", params: {dir: (.folderUri | sub("^file://"; ""))}}},
            {title: "Open Folder", key: "o", onAction: { type: "open", url: .folderUri, exit: true}}
        ]
    }' | sunbeam query -s '{
        type: "list",
        items: .
    }'
elif [ "$COMMAND" = "open-project" ]; then
    dir=$(sunbeam query -r '.params.dir' <<< "$1")
    code "$dir"
fi


