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
                        "type": "runScript",
                        "script": "writeEntry",
                        "title": "Write Entry",
                        "reloadPage": True,
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
                        "type": "copyText",
                        "text": entry["content"],
                        "title": "Copy Message",
                    },
                    {
                        "type": "runScript",
                        "title": "New Entry",
                        "script": "writeEntry",
                        "reloadPage": True,
                        "shortcut": "ctrl+n",
                        "with": {
                            "title": {
                                "type": "textfield",
                                "title": "Title",
                            },
                            "content": {
                                "type": "textarea",
                                "title": "Content",
                            },
                        },
                    },
                    {
                        "type": "runScript",
                        "title": "Delete Entry",
                        "script": "deleteEntry",
                        "reloadPage": True,
                        "shortcut": "ctrl+d",
                        "with": {"uuid": uuid},
                    },
                    {
                        "type": "runScript",
                        "title": "Edit Entry",
                        "script": "editEntry",
                        "reloadPage": True,
                        "shortcut": "ctrl+e",
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
        )
    )
