#!/usr/bin/env python3

import sys
import json
import pathlib

if len(sys.argv) == 1:
    json.dump(
        {
            "title": "File Browser",
            "root": "ls",
            "commands": [
                {
                    "name": "ls",
                    "title": "List files",
                    "mode": "view",
                    "params": [
                        {"name": "dir", "type": "string", "optional": True},
                        {"name": "show-hidden", "type": "boolean", "optional": True},
                    ],
                }
            ],
        },
        sys.stdout,
        indent=4,
    )
    sys.exit(0)

payload = json.loads(sys.argv[1])
command = payload["command"]
params = payload["params"]

if command == "ls":
    root = pathlib.Path(params.get("dir", "."))
    show_hidden = params.get("show-hidden", False)

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
                    "onAction": {
                        "type": "run",
                        "command": "ls",
                        "params": {
                            "dir": str(file.absolute()),
                        },
                    },
                }
            )
        item["actions"].extend(
            [
                {
                    "title": "Open",
                    "key": "o",
                    "onAction": {
                        "type": "open",
                        "target": str(file.absolute()),
                        "exit": True,
                    },
                },
                {
                    "title": "Show Hidden Files" if not show_hidden else "Hide Hidden Files",
                    "key": "h",
                    "onAction": {
                        "type": "reload",
                        "params": {
                            "show-hidden": not show_hidden,
                            "dir": str(root.absolute()),
                        },
                    },
                },
            ]
        )

        items.append(item)

    json.dump(
        {
            "type": "list",
            "items": items,
        },
        sys.stdout,
    )
