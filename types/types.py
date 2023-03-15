from typing import TypedDict, Literal, NotRequired
import json, sys


class List(TypedDict):
    type: Literal["list"]
    title: NotRequired[str]
    items: list["ListItem"]
    actions: NotRequired[list["Action"]]


class Detail(TypedDict):
    type: Literal["detail"]
    title: NotRequired[str]


class ListItem(TypedDict):
    title: str
    subtitle: NotRequired[str]
    actions: NotRequired[list["Action"]]


class RunAction(TypedDict):
    type: Literal["run"]
    title: str
    shortcut: NotRequired[str]
    command: str


class EditAction(TypedDict):
    type: Literal["edit"]


class ReadAction(TypedDict):
    type: Literal["read"]


class OpenAction(TypedDict):
    type: Literal["open"]


class CopyAction(TypedDict):
    type: Literal["copy"]


Action = RunAction
Page = List | Detail


hi: Page = {
    "type": "list",
    "items": [
        {"title": "hi", "actions": [{"type": "run", "title": "hi", "command": "echo"}]}
    ],
}

json.dump(hi, sys.stdout, indent=2)
