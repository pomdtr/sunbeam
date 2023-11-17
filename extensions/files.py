#!/usr/bin/env python3

import sys
import json
import pathlib

if len(sys.argv) == 1:
    home = str(pathlib.Path.home().absolute())
    json.dump(
        {
            "title": "File Browser",
            "description": "Browse files and folders",
            "items": [
                {
                    "title": "Browse Home Directory",
                    "command": "ls",
                    "params": {
                        "dir": home,
                    },
                },
                {
                    "title": "Browse Current Directory",
                    "command": "ls"
                },
                {
                    "title": "Browse Root Directory",
                    "command": "ls",
                    "params": {
                        "dir": "/",
                    },
                }
            ],
            "commands": [
                {
                    "name": "ls",
                    "title": "List files",
                    "mode": "list",
                    "params": [
                        {"name": "dir", "title": "Directory", "type": "text", "required": False}
                    ],
                }
            ],
        },
        sys.stdout,
        indent=4,
    )
    sys.exit(0)


payload = json.loads(sys.argv[1])
if payload["command"] == "ls":
    params = payload.get("params", {})
    directory = params.get("dir", payload["cwd"])
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)
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
                    "target": str(file.absolute()),
                    "exit": True,
                }
            ]
        )

        items.append(item)

    json.dump(
        {
            "items": items,
        },
        sys.stdout,
    )
