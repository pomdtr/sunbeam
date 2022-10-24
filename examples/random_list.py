#!/usr/bin/env python3

import json
import random

for _ in range(5):
    print(
        json.dumps(
            {
                "title": "Random number {}".format(random.randint(0, 100)),
                "actions": [
                    {
                        "type": "reload",
                    }
                ],
            }
        )
    )
