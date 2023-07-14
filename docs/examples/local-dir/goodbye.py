#!/usr/bin/env python3

# @sunbeam.title Say Goodbye

import json
import sys

json.dump({
    "type": "detail",
    "text": "Goodbye, world!"
}, sys.stdout)
