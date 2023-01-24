#!/usr/bin/env python3

import json, sys
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
                        "command": "write-entry",
                        "title": "Write Entry",
                        "onSuccess": "reload-page",
                        "with": {
                            "title": {"type": "textfield", "title": "Title"},
                            "content": {"type": "textarea", "title": "Content"},
                        },
                    }
                ],
            }
        )
    )
    exit()

json.dump(
    {
        "type": "list",
        "items": [
            {
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
                        "command": "write-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+n",
                        "with": {
                            "title": {"type": "textfield", "title": "Title"},
                            "content": {"type": "textarea", "title": "Content"},
                        }
                    },
                    {
                        "type": "run-command",
                        "title": "Delete Entry",
                        "command": "delete-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+d",
                        "with": {"uuid": uuid},
                    },
                    {
                        "type": "run-command",
                        "title": "Edit Entry",
                        "command": "edit-entry",
                        "shortcut": "ctrl+e",
                        "onSuccess": "reload-page",
                        "with": {
                            "uuid": uuid,
                            "title": {
                                "type": "textfield",
                                "title": "Title",
                                "default": entry["title"],
                            },
                            "content": {
                                "type": "textarea",
                                "title": "Content",
                                "default": entry["content"],
                            },
                        },
                    },
                ],
            }
            for uuid, entry in journal["entries"].items()
        ],
    },
    sys.stdout,
)
