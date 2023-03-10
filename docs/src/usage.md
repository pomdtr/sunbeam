# Usage

## Running Sunbeam

To run sunbeam, you need to provide a command as the first argument. Sunbeam will run the command and interpret it's output as a page.

```bash
sunbeam -- file-browser.py
```

Once the output is shown, you can use the arrow keys to navigate the UI, and press enter to select an item.
Use the `tab` key to show all available actions for the selected item.

Depending on the action, sunbeam may:

- Open an URI in the default application/browser
- Copy text to the system clipboard
- Run a command, and display its output on stdout
- Reload the current page
- Push a new page to the navigation stack
- Pop the current page from the navigation stack (escape key)

## Extending Sunbeam

Sunbeam is designed to be extended. You can write your own commands in any language, as long as they can output JSON.

Basically, a sunbeam script is written like a classic CLI app. However, instead of writing human-readable text to stdout, it writes JSON payload following the [Sunbeam JSON Schema](./schema.md).

### Writing a static list

The simplest sunbeam script you can write is a json file, that describe a static list.

```json
{
  "title": "My first sunbeam script",
  "type": "list",
  "items": [
    {
      "title": "Item 1",
      "subtitle": "Subtitle 1"
    },
    {
      "title": "Item 2",
      "subtitle": "Subtitle 2"
    }
  ]
}
```

Create a file named `sunbeam.json` and run `sunbeam read sunbeam.json` to show the list.

### Writing a dynamic list

However, sunbeam scripts can be much more powerful. They can be used to generate dynamic lists, or to interact with external services.

Let's generate the list of files in the current directory.

```python
#!/usr/bin/env python3

import json
from pathlib import Path

items = [
    {
        "title": path.name,
        "subtitle": str(path),
        "actions": [
            {
                "type": "open",
                "title": "Open File",
                "target": str(path),
            },
            {
                "type": "copy",
                "title": "Copy Path",
                "shortcut": "ctrl+y",
                "text": str(path),
            },
        ],
    }
    for path in Path(".").iterdir()
]

print(json.dumps({"type": "list", "items": items}))
```

Save this script as `file-browser.py`, make it executable using `chmod +x ./file-browser.py` and run `sunbeam run ./file-browser.py` to show the list.

### Adding Arguments

Let's add some options to our script to make it more useful.

```python
#!/usr/bin/env python3

import argparse
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("root", nargs="?", type=str, default=".")
parser.add_argument("--show-hidden", action="store_true")
args = parser.parse_args()

root: pathlib.Path = pathlib.Path(args.root)
entries = root.iterdir()
if not args.show_hidden:
    entries = filter(lambda p: not p.name.startswith("."), entries)



items = [
    {
        "title": filepath.name,
        "subtitle": str(filepath),
        "actions": [
            {
                "type": "open",
                "title": "Open File",
                "target": str(filepath),
            },
            {
                "type": "copy",
                "title": "Copy Filepath",
                "shortcut": "ctrl+y",
                "text": str(filepath),
            },
        ],
    }
    for filepath in entries
]

print(json.dumps({"type": "list", "items": items}))
```

You can now run `sunbeam run file-browser.py /tmp` to show the list of files in the `/tmp` directory, or `sunbeam run -- file-browser.py --show-hidden` to show hidden files in the current directory.

Notice that we used the `--` argument separator to pass arguments to the script. This is required because sunbeam also accepts flags, and we don't want it to interpret them.

### Adding Navigation

This is nice, but we can do better. A full-blown file browser would allow us to navigate through directories.
In sunbeam, navigate to a new page using the `push` action.

This schema describe the sunbeam event loop. Each time a `push` action is triggered, a new command is run, and the output is pushed to the navigation stack.

![Sunbeam Event Loop](./assets/event-loop.excalidraw.png)

Let's update our file browser to support navigation.

```python
{{#include ./examples/file-browser/file-browser.py}}
```

### What's next?

This is just the tip of the iceberg. Sunbeam can show detail pages, prompt the user for input, and much more.

To learn more, check out the [Sunbeam JSON Schema](./schema.md) and the provided [examples](./examples).

In order to run the examples, just clone the sunbeam repository and run `sunbeam` from the root directory.

> **Warning** Some examples require external dependencies. Please refer to the README of each example for more information.
