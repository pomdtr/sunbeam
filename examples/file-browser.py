#!/usr/bin/env python3

import json
import pathlib
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--root", type=pathlib.Path)
args = parser.parse_args()

root: pathlib.Path = args.root

for path in root.iterdir():
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
                        "type": "copy-content",
                        "title": "Copy Path",
                        "content": str(path.absolute()),
                    },
                ],
            }
        )
    )
