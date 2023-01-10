#!/usr/bin/env python3

import json
import pathlib
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--root", required=True, type=pathlib.Path)
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root: pathlib.Path = args.root

entries = root.iterdir()
if not args.show_hidden:
    entries = filter(lambda p: not p.name.startswith("."), entries)

for path in sorted(entries, key=lambda p: p.name):
    primaryAction = (
        {"type": "open-path", "path": str(path.absolute()), "title": "Open File"}
        if path.is_file()
        else {
            "type": "run-script",
            "script": "browseFiles",
            "title": "Browse Directory",
            "with": {"root": str(path.absolute()), "showHidden": args.show_hidden},
        }
    )

    print(
        json.dumps(
            {
                "title": path.name,
                "accessories": [str(root.absolute())],
                "actions": [
                    primaryAction,
                    {
                        "type": "copy-text",
                        "title": "Copy Path",
                        "shorcut": "ctrl+y",
                        "text": str(path.absolute()),
                    },
                    {
                        "type": "reload-page",
                        "title": "Toggle Hidden Files",
                        "shortcut": "ctrl+h",
                        "with": {"showHidden": not args.show_hidden},
                    }
                ],
            }
        )
    )
