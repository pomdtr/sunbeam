#!/usr/bin/env python3

import json
from journal import load_journal
from datetime import datetime


journal = load_journal()

if len(journal["entries"]) == 0:
    print(
        json.dumps(
            {
                "title": "No entries",
                "preview": "Hit enter to create a new entry !",
                "actions": [
                    {
                        "type": "run-command",
                        "script": "write-entry",
                        "title": "Write Entry",
                        "silent": True,
                        "onSuccess": "reload-page",
                        "with": {
                            "title": {
                                "type": "textfield",
                                "title": "Title",
                            },
                            "content": {
                                "type": "textfield",
                                "title": "Content",
                            },
                        },
                    }
                ],
            }
        )
    )
    exit()

for uuid, entry in journal["entries"].items():
    print(
        json.dumps(
            {
                "id": uuid,
                "title": entry["title"],
                "preview": entry["content"],
                "accessories": [
                    datetime.utcfromtimestamp(entry["timestamp"]).strftime(
                        "%Y-%m-%d %H:%M:%S"
                    )
                ],
                "actions": [
                    {
                        "type": "copy-text",
                        "text": entry["content"],
                        "title": "Copy Message",
                    },
                    {
                        "type": "run-command",
                        "title": "New Entry",
                        "script": "writeEntry",
                        "onSuccess": "reload-page",
                        "silent": True,
                        "shortcut": "ctrl+n"
                    },
                    {
                        "type": "run-command",
                        "title": "Delete Entry",
                        "script": "deleteEntry",
                        "onSuccess": "reload-page",
                        "silent": True,
                        "shortcut": "ctrl+d",
                        "with": {"uuid": uuid},
                    },
                    {
                        "type": "run-command",
                        "title": "Edit Entry",
                        "script": "editEntry",
                        "silent": True,
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+e",
                        "with": {
                            "uuid": uuid,
                            "title": {
                                "type": "textfield",
                                "title": "Title",
                                "required": True,
                                "default": entry["title"],
                            },
                            "content": {
                                "type": "textarea",
                                "required": True,
                                "title": "Content",
                                "default": entry["content"],
                            },
                        },
                    },
                ],
            }
        )
    )
