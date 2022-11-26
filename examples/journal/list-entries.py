#!/usr/bin/env python3

import json
from journal import load_journal


journal = load_journal()

if len(journal['entries']) == 0:
    print(json.dumps({
        'title': 'No entries',
        'actions': [
            {
                'type': 'run-script',
                'script': 'write-entry',
                'title': 'Write Entry',
            }
        ]
    }))
    exit()

for index, entry in enumerate(journal["entries"]):
    print(
        json.dumps(
            {
                "title": entry["title"],
                "preview": entry["message"],
                "accessories": [entry["date"]],
                "actions": [
                    {
                        "type": "copy-text",
                        "text": entry["message"],
                        "title": "Copy Message",
                    },
                    {
                        "type": "run-script",
                        "title": "New Entry",
                        "script": "write-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+n",

                    },
                    {
                        "type": "run-script",
                        "title": "Delete Entry",
                        "script": "delete-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+d",
                        "with": {"index": str(index)},
                    },
                ],
            }
        )
    )
