#!/usr/bin/env python3

# @title Say Goodbye

import json
import sys

json.dump({
    "type": "detail",
    "text": "Goodbye, world!"
}, sys.stdout)
