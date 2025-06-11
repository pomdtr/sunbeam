#!/usr/bin/env python3

import sys
import json
import pathlib

if len(sys.argv) == 1:
    json.dump(
        {
            "title": "File Browser",
            "description": "Browse files and folders",
            "root": [
                {
                    "title": "Browse Home Directory",
                    "type": "run",
                    "command": "browse",
                    "params": {
                        "dir": "~",
                    },
                },
                {
                    "title": "Browse Current Directory",
                    "type": "run",
                    "command": "browse",
                    "params": {
                        "dir": ".",
                    },
                },
            ],
            "commands": [
                {
                    "name": "browse",
                    "mode": "filter",
                    "params": [
                        {
                            "name": "dir",
                            "type": "string",
                            "optional": True,
                        },
                    ],
                }
            ],
        },
        sys.stdout,
        indent=4,
    )
    sys.exit(0)

if sys.argv[1] == "browse":
    # Parse CLI arguments
    params = {}
    i = 2
    while i < len(sys.argv):
        if sys.argv[i].startswith("--"):
            key = sys.argv[i][2:]
            if i + 1 < len(sys.argv):
                params[key] = sys.argv[i + 1]
                i += 2
            else:
                i += 1
        else:
            i += 1
    
    directory = params.get("dir", ".")
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)

    items = []
    for file in root.iterdir():
        if file.name.startswith("."):
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
                    "command": "browse",
                    "params": {
                        "dir": str(file.absolute()),
                    },
                }
            )

        item["actions"].extend(
            [{"title": "Open", "type": "open", "target": str(file.absolute())}]
        )

        items.append(item)

    print(json.dumps({"items": items}))
