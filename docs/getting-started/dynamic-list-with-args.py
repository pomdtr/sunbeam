#!/usr/bin/env python3

import argparse
import json
import sys
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("root", nargs="?", type=str, default=".")
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root = Path(args.root)
entries = root.iterdir()
if not args.show_hidden:
    entries = filter(lambda p: not p.name.startswith("."), entries)


items = [
    {
        "title": filepath.name,
        "subtitle": str(filepath),
        "actions": [
            {
                "type": "open",
                "title": "Open File",
                "path": str(filepath),
            },
            {
                "type": "copy",
                "title": "Copy Filepath",
                "key": "y",
                "text": str(filepath),
            },
        ],
    }
    for filepath in entries
]

json.dump({"type": "list", "items": items}, sys.stdout)
