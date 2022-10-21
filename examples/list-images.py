#!/usr/bin/env python3

#  @sunbeam.schemaVersion 1
#  @sunbeam.title List Images
#  @sunbeam.mode interactive
#  @sunbeam.subtitle Docker

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
        "actions": [{"type": "copy"}],
    }
    for image in images
]
for item in items:
    print(json.dumps(item))
