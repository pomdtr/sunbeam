#!/usr/bin/env python3

import json
from pathlib import Path

items = [
    {
        "title": path.name,
        "subtitle": str(path),
        "actions": [
            {
                "type": "open",
                "title": "Open File",
                "target": str(path),
            },
            {
                "type": "copy",
                "title": "Copy Path",
                "shortcut": "ctrl+y",
                "text": str(path),
            },
        ],
    }
    for path in Path(".").iterdir()
]

print(json.dumps({"type": "list", "items": items}))
