#!/usr/bin/env python3

import argparse
import json
import pathlib
from typing import TypedDict
import sys
import uuid


class TodoItem(TypedDict):
    title: str
    done: bool


class TodoList(TypedDict):
    items: dict[str, TodoItem]


todolist_path = pathlib.Path("/tmp/todolist.json")


def load_todolist() -> TodoList:
    if not todolist_path.exists():
        return {"items": {}}

    with open(todolist_path) as f:
        return json.load(f)


def save_todolist(todolist: TodoList):
    with open(todolist_path, "w") as f:
        json.dump(todolist, f)


def handle(args: argparse.Namespace):
    todolist = load_todolist()

    if args.command == "list" or args.command is None:
        print(
            json.dumps(
                {
                    "type": "list",
                    "actions": [
                        {
                            "type": "run",
                            "title": "Add Item",
                            "command": f"{sys.argv[0]} add ${{input:title}}",
                            "inputs": [
                                {
                                    "name": "title",
                                    "title": "Title",
                                    "type": "textfield",
                                }
                            ],
                            "shortcut": "ctrl+n",
                            "onSuccess": "reload",
                        }
                    ],
                    "items": [
                        {
                            "id": key,
                            "title": item["title"],
                            "subtitle": "Done" if item["done"] else "Todo",
                            "accessories": [key],
                            "actions": [
                                {
                                    "type": "run",
                                    "title": "Toggle Completion",
                                    "command": f"{sys.argv[0]} toggle {key}",
                                    "onSuccess": "reload",
                                },
                                {
                                    "type": "run",
                                    "title": "Edit Title",
                                    "shortcut": "ctrl+e",
                                    "onSuccess": "reload",
                                    "command": f"{sys.argv[0]} edit-title {key} ${{input:title}}",
                                    "inputs": [
                                        {
                                            "name": "title",
                                            "title": "Title",
                                            "type": "textfield",
                                        }
                                    ],
                                },
                                {
                                    "type": "run",
                                    "title": "Delete",
                                    "shortcut": "ctrl+d",
                                    "onSuccess": "reload",
                                    "command": f"{sys.argv[0]} delete {key}",
                                },
                            ],
                        }
                        for key, item in todolist["items"].items()
                    ],
                }
            )
        )

    elif args.command == "add":
        key = str(uuid.uuid4())
        todolist.setdefault("items", {})[key] = {"title": args.title, "done": False}
        save_todolist(todolist)

    elif args.command == "edit-title":
        key = args.key
        todolist["items"][key]["title"] = args.title
        save_todolist(todolist)

    elif args.command == "toggle":
        key = args.key
        todolist["items"][key]["done"] = not todolist["items"][key]["done"]
        print(todolist["items"][key]["done"])
        save_todolist(todolist)

    elif args.command == "delete":
        key = args.key
        del todolist["items"][key]
        save_todolist(todolist)

    else:
        raise ValueError(f"Unknown command {args.command}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command")

    list = subparsers.add_parser("list")

    add = subparsers.add_parser("add")
    add.add_argument("title", type=str)

    toggle = subparsers.add_parser("toggle")
    toggle.add_argument("key", type=str)

    edit_title = subparsers.add_parser("edit-title")
    edit_title.add_argument("key", type=str)
    edit_title.add_argument("title", type=str)

    delete = subparsers.add_parser("delete")
    delete.add_argument("key", type=str)

    args = parser.parse_args()

    handle(args)
