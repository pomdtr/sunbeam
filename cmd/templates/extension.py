#!/usr/bin/env python3

import sys
import json

if len(sys.argv) < 2:
    json.dump({
        "title": "Example Extension",
        "description": "Example extension",
        "commands": [
            {
                "name": "hello",
                "title": "Hello",
                "mode": "detail"
            }
        ]
    }, sys.stdout)
    sys.exit(0)

payload = json.loads(sys.argv[1])
if payload["command"] == "hello":
    json.dump({"text": "Hello, world!"}, sys.stdout)
