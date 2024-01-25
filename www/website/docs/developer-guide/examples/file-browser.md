# File Browser

## Writing the manifest

If the script is called without arguments, it must return a json manifest describing the extension and its commands.

```python
#!/usr/bin/env python3

import sys
import json
import pathlib

manifest = {
    "title": "File Browser",
    "description": "Browse files and folders",
    "preferences": [
        {
            "name": "show-hidden",
            "title": "Show Hidden Files",
            "type": "boolean",
            "optional": True
        }
    ],
    "commands": [
        {
            "name": "ls",
            "title": "List files",
            "mode": "filter",
            "params": [
                {
                    "name": "dir",
                    "title": "Directory",
                    "type": "text",
                    "optional": True,
                },
            ],
        }
    ],
}


if len(sys.argv) == 1:
    print(json.dumps(manifest))
    sys.exit(0)
```

## Writing the command

We want to be able to browse the filesystem, so the `ls` command is recursive. It will list the files in the current directory, and call itself again if the user wants to browse a directory.

```python
#!/usr/bin/env python3

import sys
import json
import pathlib

manifest = {
    "title": "File Browser",
    "description": "Browse files and folders",
    "preferences": [
        {
            "name": "show-hidden",
            "title": "Show Hidden Files",
            "type": "boolean",
            "optional": True,
        }
    ],
    "commands": [
        {
            "name": "ls",
            "title": "List files",
            "mode": "filter",
            "params": [
                {
                    "name": "dir",
                    "title": "Directory",
                    "type": "string",
                    "optional": True,
                },
            ],
        }
    ],
}

if len(sys.argv) == 1:
    json.dump(
        manifest,
        sys.stdout,
        indent=4,
    )
    sys.exit(0)

payload = json.loads(sys.argv[1])
params = payload["params"]
preferences = payload["preferences"]
cwd = payload["cwd"]

if payload["command"] == "ls":
    directory = params["dir"] or cwd
    if directory.startswith("~"):
        directory = directory.replace("~", str(pathlib.Path.home()))
    root = pathlib.Path(directory)
    show_hidden = preferences.get("show-hidden", False)

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
                    "path": str(file.absolute()),
                },
                {
                    "title": "Show Hidden Files"
                    if not show_hidden
                    else "Hide Hidden Files",
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
else:
    print(f"Unknown command: {payload['command']}")
    sys.exit(1)
```
