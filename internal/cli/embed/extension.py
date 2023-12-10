#!/usr/bin/env python3

import sys
import json

if len(sys.argv) == 1:
    manifest = {
        "title": "My Extension",
        "description": "This is my extension",
        "items": [
            {
                "title": "Hi Mom!",
                "command": "hi",
                "params": {
                    "name": "Mom"
                }
            },
            {
                "command": "hi",
            }
        ],
        "commands": [
            {
                "name": "hi",
                "title": "Say Hi",
                "mode": "detail",
                "params": [
                    {
                        "name": "name",
                        "title": "Name",
                        "type": "text"
                    }
                ]
            }
        ]
    }
    print(json.dumps(manifest))
    sys.exit(0)

payload = json.loads(sys.argv[1])

if payload["command"] == "hi":
    name = payload["params"]["name"]
    detail = {
        "text": f"Hi {name}!",
        "actions": [
            {
                "title": "Copy Name",
                "type": "copy",
                "text": name
            }
        ]
    }
    print(json.dumps(detail))
else:
    print("Unknown command")
    sys.exit(1)
