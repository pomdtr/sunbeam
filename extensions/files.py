#!/usr/bin/env python3

import sys
import json
import pathlib

manifest = {
    "title": "File Browser",
    "description": "Browse files and folders",
    "commands": [
        {
            "name": "ls",
            "title": "List files",
            "mode": "filter",
            "params": [
                {
                    "name": "dir",
                    "title": "Directory",
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
    params = payload["params"]
    preferences = payload["preferences"]
    directory = params["dir"] or payload["cwd"]
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)

    items = []
    for file in root.iterdir():
        item = {
            "title": file.name,
            "accessories": [str(file.absolute())],
            "actions": [],
        }
        if file.is_dir():
            item["actions"].append(
                {
                    "title": "Browse",
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
                    "extension": "std",
                    "command": "open",
                    "params": {
                        "url": f"file:{str(file.absolute())}",
                    },
                },
            ]
        )

        items.append(item)

    print(json.dumps({"items": items}))
