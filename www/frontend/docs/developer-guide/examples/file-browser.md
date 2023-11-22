## File Browser

This scripts is a good examples on how to handle navigation in sunbeam. Each time the user navigates to a new directory, the script is rerun with the new directory as the `dir` parameter.

If no directory is provided, the script will list the content of the current working directory, which is provided by sunbeam as the `cwd` field of the payload.

```python
#!/usr/bin/env python3

import sys
import json
import pathlib

if len(sys.argv) == 1:
    json.dump(
        {
            "title": "File Browser",
            "description": "Browse files and folders",
            "root": ["ls"],
            "commands": [
                {
                    "name": "ls",
                    "title": "List files",
                    "mode": "list",
                    "params": [
                        {"name": "dir", "description": "Directory", "type": "string"},
                        {"name": "show-hidden", "description": "Show Hidden Files", "type": "boolean"},
                    ],
                }
            ],
        },
        sys.stdout,
        indent=4,
    )
    sys.exit(0)


payload = json.loads(sys.argv[1])
if payload["command"] == "ls":
    params = payload.get("params", {})
    directory = params.get("dir", payload["cwd"])
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)
    show_hidden = params.get("show-hidden", False)

    items = []
    for file in root.iterdir():
        if not show_hidden and file.name.startswith("."):
            continue
        item = {
            "title": file.name,
            "accessories": [str(file.absolute())],
            "actions": [],
        }
        if file.is_dir():
            item["actions"].append(
                {
                    "title": "Browse",
                    "type": "run",
                    "command": "ls",
                    "params": {
                        "dir": str(file.absolute()),
                    },
                }
            )
        item["actions"].extend(
            [
                {
                    "title": "Open",
                    "key": "o",
                    "type": "open",
                    "target": str(file.absolute()),
                    "exit": True,
                },
                {
                    "title": "Show Hidden Files" if not show_hidden else "Hide Hidden Files",
                    "key": "h",
                    "type": "reload",
                    "params": {
                        "show-hidden": not show_hidden,
                        "dir": str(root.absolute()),
                    },
                },
            ]
        )

        items.append(item)

    print(json.dumps({"items": items}))
```

You can pin a specificic directory by adding an item to the `oneliner` array in the sunbeam config:

```json
{
    "oneliners": [
        {
            "command": "sunbeam file-browser list --dir ~/Downloads",
            "title": "Search Downloads",
        }
    ]
}
```
