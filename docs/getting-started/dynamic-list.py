#!/usr/bin/env python3

import json
from pathlib import Path

items = [
    {
        "title": path.name,
        "subtitle": str(path),
        "actions": [
            {
                "type": "open-path",
                "title": "Open File",
                "path": str(path),
            },
            {
                "type": "copy-text",
                "title": "Copy Path",
                "key": "y",
                "text": str(path),
            },
        ],
    }
    for path in Path(".").iterdir()
]

print(json.dumps({"type": "list", "items": items}))
