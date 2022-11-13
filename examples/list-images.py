#!/usr/bin/env python3

import subprocess
import json
import sys

try:
    res = subprocess.run(
        ["docker", "image", "ls", "--format", "{{ json . }}"],
        text=True,
        stderr=subprocess.PIPE,
        stdout=subprocess.PIPE,
        check=True,
    )
except subprocess.CalledProcessError as e:
    print(e.stderr, file=sys.stderr)
    sys.exit(1)

images = [json.loads(line) for line in res.stdout.splitlines()]

items = [
    {
        "title": image["Repository"],
        "subtitle": image["Tag"],
        "actions": [
            {"type": "copy-content", "content": image["Repository"]},
            {"type": "copy-content", "content": image["Repository"]},
            {
                "type": "run-script",
                "script": "delete-image",
                "with": {"image": image["Repository"]},
            },
        ],
    }
    for image in images
]
for item in items:
    print(json.dumps(item))
