#!/usr/bin/env python3

# @title Say Hello

import json
import sys

json.dump({
    "type": "detail",
    "text": "Hello, world!"
}, sys.stdout)
