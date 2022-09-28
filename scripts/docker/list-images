#!/usr/bin/env python3

#  @sunbeam.schemaVersion 1
#  @sunbeam.title List Images
#  @sunbeam.packageName Docker

import subprocess
import json

res = subprocess.run(
    ["docker", "image", "ls", "--format", "{{ json . }}"],
    text=True,
    capture_output=True,
    check=True,
)

images = [json.loads(line) for line in res.stdout.splitlines()]

view = {
    "type": "list",
    "list": {
        "items": [
            {"title": image["Repository"], "subtitle": image["Tag"]} for image in images
        ]
    },
}

print(json.dumps(view))
