#!/usr/bin/env python3

#  @sunbeam.schemaVersion 1
#  @sunbeam.title List Images
#  @sunbeam.mode interactive
#  @sunbeam.packageName Docker

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

view = {
    "type": "list",
    "list": {
        "items": [
            {"title": image["Repository"], "subtitle": image["Tag"]} for image in images
        ]
    },
}

print(json.dumps(view))
