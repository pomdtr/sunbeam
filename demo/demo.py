#!/usr/bin/env python3

import json

page = {
    "type": "list",
    "title": "Sunbeam Features",
    "showPreview": True,
    "items": [
        {
            "title": "Runs Everywhere",
            "preview": {
                "type": "static",
                "language": "markdown",
                "text": """Sunbeam is a cross-platform application that runs on **Windows**, **macOS**, and **Linux**.

It is distributed as a single binary that you can download from github releases, or install with a package manager.

```bash
# Install with Homebrew
brew install pomdtr/tap/sunbeam

# Install with Scoop
scoop install pomdtr/tap/sunbeam

# Install from source
go install github.com/pomdtr/sunbeam@latest
```
"""
            }
        },
        {
            "title": "Supports any language",
            "preview": {
                "type": "static",
                "language": "markdown",
                "text": """Sunbeam supports any language. It is the the most bare-bone UI framework you can imagine.

There is no complex API to learn, just write a script that outputs JSON and Sunbeam will take care of the rest.

This is a valid sunbeam script:

```bash
echo '{"type": "list", "items": [{"title": "Hello World!"}]}'
```
"""
            }
        },
        {
            "title": "Github Extensions",
            "preview": {
                "type": "static",
                "language": "markdown",
                "text": """Any Github repository can become a Sunbeam extension. Just create an executable `sunbeam-extension` script at the root of your repository.

Alternatively, you can compile your extension into a binary and add publish it as a release.

If you want to publish your extension, just add the sunbeam-extension topic to your repository.

Any user can install your extension by running `sunbeam install <username>/<repository>`.
"""
            }
        }
    ]
}

print(json.dumps(page, indent=2))
