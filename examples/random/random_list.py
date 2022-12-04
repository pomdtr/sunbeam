#!/usr/bin/env python3

import json
import random

for _ in range(5):
    random_number = random.randint(0, 100)
    print(
        json.dumps(
            {
                "title": f"Random number {random_number}",
                "actions": [
                    {
                        "title": "Reload List",
                        "shortcut": "enter",
                        "type": "reloadPage",
                    },
                    {
                        "title": "Copy",
                        "shortcut": "ctrl+y",
                        "type": "copyText",
                        "text": str(random_number),
                    },
                ],
            }
        )
    )
