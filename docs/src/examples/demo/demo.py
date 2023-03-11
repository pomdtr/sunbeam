#!/usr/bin/env python3

import argparse
import json
import sys

parser = argparse.ArgumentParser()

subparsers = parser.add_subparsers(dest="command")

static_list = subparsers.add_parser("static-list")
static_list.add_argument("nb_items", type=int)

dynamic_list = subparsers.add_parser("dynamic-list")

help = subparsers.add_parser("help")

args = parser.parse_args()

welcome_text = """
Welcome to sunbeam !

Hit tab to see the list of available actions.
"""

help_page = """
I'm a command-line launcher, inspired by [fzf](https://github.com/junegunn/fzf), [raycast](https://raycast.com) and [gum](https://github.com/charmbracelet/gum).

I support:

- Running on all platforms (Windows, Linux, MacOS)
- Commands written in any language, as long as they can output JSON
- Generating powerful UIs composed of a succession of forms, lists, details views

Now hit escape to go back to the previous view.
"""

if not args.command:
    json.dump(
        {
            "type": "detail",
            "content": {
                "text": welcome_text,
            },
            "actions": [
                {
                    "type": "run",
                    "onSuccess": "push",
                    "title": "Tell me more about you!",
                    "command": sys.argv[0],
                    "args": ["help"],
                },
                {
                    "type": "run",
                    "onSuccess": "push",
                    "title": "Show me a static list!",
                    "command": sys.argv[0],
                    "args": ["static-list", "${input:nb_items}"],
                    "inputs": [
                        {
                            "name": "nb_items",
                            "type": "textfield",
                            "title": "Number of items to show",
                        }
                    ],
                },
                {
                    "type": "run",
                    "onSuccess": "push",
                    "title": "Show me a dynamic list!",
                    "command": sys.argv[0],
                    "args": ["dynamic-list"],
                },
                {
                    "type": "open",
                    "title": "Open the Docs",
                    "target": "https://pomdtr.github.io/sunbeam/",
                },
                {
                    "type": "copy",
                    "title": "Copy list alias",
                    "text": " ".join(
                        ["sunbeam", "run", sys.argv[0], "static-list", "5"]
                    ),
                },
            ],
        },
        sys.stdout,
    )
elif args.command == "help":
    json.dump(
        {
            "type": "detail",
            "content": {
                "text": help_page,
            },
        },
        sys.stdout,
    )
elif args.command == "dynamic-list":
    query = sys.stdin.read()
    json.dump(
        {
            "type": "list",
            "generateItems": True,
            "items": [
                {
                    "title": query.upper(),
                    "subtitle": "Upper case",
                    "actions": [
                        {
                            "type": "copy",
                            "text": query.upper(),
                        }
                    ],
                },
                {
                    "title": query.lower(),
                    "subtitle": "Lower case",
                    "actions": [
                        {
                            "type": "copy",
                            "text": query.lower(),
                        }
                    ],
                },
                {
                    "title": query.title(),
                    "subtitle": "Title case",
                    "actions": [
                        {
                            "type": "copy",
                            "text": query.title(),
                        }
                    ],
                },
                {
                    "title": query.capitalize(),
                    "subtitle": "Capitalize",
                    "actions": [
                        {
                            "type": "copy",
                            "text": query.capitalize(),
                        }
                    ],
                },
                {
                    "title": query.swapcase(),
                    "subtitle": "Swap case",
                    "actions": [
                        {
                            "type": "copy",
                            "text": query.swapcase(),
                        }
                    ],
                },
            ]
            if query
            else [{"title": "Please enter a query"}],
        },
        sys.stdout,
    )
elif args.command == "static-list":
    json.dump(
        {
            "type": "list",
            "showPreview": True,
            "items": [
                {
                    "title": "Title",
                    "subtitle": "Subtitle",
                    "accessories": [f"Item {i+1}"],
                    "preview": {
                        "text": f"Preview text of item {i+1}",
                    },
                    "actions": [],
                }
                for i in range(args.nb_items)
            ],
        },
        sys.stdout,
    )
