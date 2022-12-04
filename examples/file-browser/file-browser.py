#!/usr/bin/env python3

import json
import pathlib
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--root", required=True, type=pathlib.Path)
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root: pathlib.Path = args.root

items = sorted(root.iterdir(), key=lambda p: p.name)
if not args.show_hidden:
    items = filter(lambda p: not p.name.startswith("."), items)

for path in sorted(root.iterdir(), key=lambda p: p.name):
    primaryAction = (
        {"type": "openPath", "path": str(path.absolute())}
        if path.is_file()
        else {
            "type": "runScript",
            "script": "browseFiles",
            "title": "Browse Directory",
            "with": {"root": str(path.absolute())},
        }
    )

    print(
        json.dumps(
            {
                "title": path.name,
                "subtitle": str(root.absolute()),
                "actions": [
                    primaryAction,
                    {
                        "type": "copyText",
                        "title": "Copy Path",
                        "text": str(path.absolute()),
                    },
                ],
            }
        )
    )
