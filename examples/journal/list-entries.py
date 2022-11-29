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
                        "type": "run-script",
                        "script": "write-entry",
                        "title": "Write Entry",
                        "onSuccess": "reload-page",
                        "with": [
                            {
                                "param": "title",
                                "value": {
                                    "type": "textfield",
                                    "title": "Title",
                                },
                            },
                            {
                                "param": "content",
                                "value": {
                                    "type": "textfield",
                                    "title": "Content",
                                },
                            },
                        ],
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
                        "type": "run-script",
                        "title": "New Entry",
                        "script": "write-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+n",
                        "with": [
                            {
                                "param": "title",
                                "value": {"type": "textfield", "title": "Title"},
                            },
                            {
                                "param": "content",
                                "value": {"type": "textarea", "title": "Content"},
                            },
                        ],
                    },
                    {
                        "type": "run-script",
                        "title": "Delete Entry",
                        "script": "delete-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+d",
                        "with": [{"param": "uuid", "value": uuid}],
                    },
                    {
                        "type": "run-script",
                        "title": "Edit Entry",
                        "script": "edit-entry",
                        "onSuccess": "reload-page",
                        "shortcut": "ctrl+e",
                        "with": [
                            {"param": "uuid", "value": uuid},
                            {
                                "param": "title",
                                "value": {
                                    "type": "textfield",
                                    "title": "Title",
                                    "value": entry["title"],
                                },
                            },
                            {
                                "param": "content",
                                "value": {
                                    "type": "textarea",
                                    "title": "Content",
                                    "value": entry["content"],
                                },
                            },
                        ],
                    },
                ],
            }
        )
    )
