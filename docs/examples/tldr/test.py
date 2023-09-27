#!/usr/bin/env python3

import sys
import json

# If no arguments are passed, return the list of commands
if len(sys.argv) == 1:
    print(json.dumps({
        "title": "tldr",
        "commands": [
            {
                "name": "tldr",
                "title": "List Tldr pages",
                "mode": "view"
            }
        ]
    }))
    sys.exit(0)

# Otherwise, parse the arguments and return the result

payload = json.loads(sys.argv[1])
command = payload["command"]

if command == "tldr":
    print(json.dumps({
        "type": "list",
        "items": [
            {"title": "Item 1"}
        ]
    }))
else:
    print(f"Unknown command: {command}", file=sys.stderr)
