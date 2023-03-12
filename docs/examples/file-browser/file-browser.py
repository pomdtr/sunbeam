#!/usr/bin/env python3

import json
import pathlib
import argparse
import sys

parser = argparse.ArgumentParser()
parser.add_argument("root", nargs="?", type=str)
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root: pathlib.Path = pathlib.Path(args.root) if args.root else pathlib.Path.cwd()
entries = root.iterdir()
if not args.show_hidden:
    entries = filter(lambda p: not p.name.startswith("."), entries)

json.dump(
    {
        "type": "list",
        "items": [
            {
                "title": path.name,
                "accessories": [str(root.absolute())],
                "actions": [
                    (
                        {
                            "type": "open",
                            "title": "Open File",
                            "target": f"file://{path.absolute()}",
                        }
                        if path.is_file()
                        else {
                            "type": "run",
                            "onSuccess": "push",
                            "title": "Browse Directory",
                            "command": f"{sys.argv[0]} {path.absolute()}",
                        }
                    ),
                    {
                        "type": "copy",
                        "title": "Copy Path",
                        "shortcut": "ctrl+y",
                        "text": str(path.absolute()),
                    },
                ],
            }
            for path in sorted(entries, key=lambda p: p.name)
        ],
    },
    sys.stdout,
)
