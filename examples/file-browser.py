#!/usr/bin/env python3

import json
import pathlib
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--root", type=pathlib.Path)
args = parser.parse_args()

root: pathlib.Path = args.root

for path in sorted(root.iterdir(), key=lambda p: p.name):
    primaryAction = (
        {"type": "open-file", "path": str(path.absolute())}
        if path.is_file()
        else {
            "type": "push-page",
            "page": "file-browser",
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
                        "type": "copy-text",
                        "title": "Copy Path",
                        "text": str(path.absolute()),
                    },
                ],
            }
        )
    )
