#!/usr/bin/env python3

import os
import sys
import json
import pathlib

if len(sys.argv) == 1:
    home = str(pathlib.Path.home().absolute())
    json.dump(
        {
            "title": "File Browser",
            "description": "Browse files and folders",
            "preferences": [
                {"name": "show-hidden", "type": "checkbox", "label": "Show Hidden Files", "required": False}
            ],
            "root": ["ls"],
            "commands": [
                {
                    "name": "ls",
                    "title": "List files",
                    "mode": "list",
                    "params": [
                        {"name": "dir", "title": "Directory", "type": "text", "default": ".", "required": False}
                    ],
                }
            ],
        },
        sys.stdout,
        indent=4,
    )
    sys.exit(0)


payload = json.loads(sys.argv[1])
show_hidden = payload["preferences"].get("show-hidden", False)
if payload["command"] == "ls":
    params = payload.get("params", {})
    directory = params.get("dir", payload["cwd"])
    if directory.startswith("~/"):
        root = pathlib.Path.joinpath(pathlib.Path.home(), directory[2:])
    elif not os.path.isabs(directory):
        root = pathlib.Path(payload["cwd"]).joinpath(directory)
    else:
        root = pathlib.Path(directory)
    items = []
    for file in root.iterdir():
        if not show_hidden and file.name.startswith("."):
            continue
        item = {
            "title": file.name,
            "subtitle": str(file.absolute()).replace(str(pathlib.Path.home()), "~"),
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
