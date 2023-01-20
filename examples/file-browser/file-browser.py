#!/usr/bin/env python3

import json
import pathlib
import argparse
import sys

parser = argparse.ArgumentParser()
parser.add_argument("--root", required=True, type=pathlib.Path)
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root: pathlib.Path = args.root
entries = root.iterdir()
if not args.show_hidden:
    entries = filter(lambda p: not p.name.startswith("."), entries)

json.dump(
    {
        "type": "list",
        "list": {
            "items": [
                {
                    "title": path.name,
                    "accessories": [str(root.absolute())],
                    "actions": [
                        (
                            {
                                "type": "open-url",
                                "path": f"file://{path.absolute()}",
                                "title": "Open File",
                            }
                            if path.is_file()
                            else {
                                "type": "run-command",
                                "command": "browse-files",
                                "title": "Browse Directory",
                                "with": {
                                    "root": str(path.absolute()),
                                    "showHidden": args.show_hidden,
                                },
                            }
                        ),
                        {
                            "type": "copy-text",
                            "title": "Copy Path",
                            "shorcut": "ctrl+y",
                            "text": str(path.absolute()),
                        },
                        {
                            "type": "run-command",
                            "title": "Delete File",
                            "shortcut": "ctrl+d",
                            "command": "delete-file",
                            "with": {"path": str(path.absolute())},
                        },
                        {
                            "type": "reload-page",
                            "title": "Toggle Hidden Files",
                            "shortcut": "ctrl+h",
                            "with": {"showHidden": not args.show_hidden},
                        },
                    ],
                }
                for path in sorted(entries, key=lambda p: p.name)
            ]
        },
    },
    sys.stdout,
)
