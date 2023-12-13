#!/usr/bin/env python3

import sys
import json
import pathlib

manifest = {
    "title": "File Browser",
    "description": "Browse files and folders",
    "preferences": [
        {
            "name": "show-hidden",
            "description": "Show Hidden Files",
            "type": "boolean",
            "default": False,
        }
    ],
    "commands": [
        {
            "name": "ls",
            "title": "List files",
            "mode": "filter",
            "params": [
                {
                    "name": "dir",
                    "description": "Directory",
                    "type": "string",
                    "optional": True,
                },
            ],
        }
    ],
}


if len(sys.argv) == 1:
    json.dump(
        manifest,
        sys.stdout,
        indent=4,
    )
    sys.exit(0)


payload = json.loads(sys.argv[1])
if payload["command"] == "ls":
    params = payload.get("params", {})
    preferences = payload.get("preferences", {})
    directory = params.get("dir", payload["cwd"])
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)
    show_hidden = preferences.get("show-hidden", False)

    items = []
    for file in root.iterdir():
        if not show_hidden and file.name.startswith("."):
            continue
        item = {
            "title": file.name,
            "accessories": [str(file.absolute())],
            "actions": [],
        }
        if file.is_dir():
            item["actions"].append(
                {
                    "title": "Browse",
                    "type": "run",
                    "command": "ls",
                    "params": {
                        "dir": str(file.absolute()),
                    },
                }
            )
        item["actions"].extend(
            [
                {
                    "title": "Open",
                    "key": "o",
                    "type": "open",
                    "path": str(file.absolute()),
                    "exit": True,
                },
                {
                    "title": "Show Hidden Files"
                    if not show_hidden
                    else "Hide Hidden Files",
                    "key": "h",
                    "type": "reload",
                    "params": {
                        "show-hidden": not show_hidden,
                        "dir": str(root.absolute()),
                    },
                },
            ]
        )

        items.append(item)

    print(json.dumps({"items": items}))
