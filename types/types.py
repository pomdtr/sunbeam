from typing import TypedDict, Literal, NotRequired
import json, sys


class List(TypedDict):
    type: Literal["list"]
    title: NotRequired[str]
    showDetail: NotRequired[bool]
    generateItems: NotRequired[bool]
    emptyText: NotRequired[str]
    items: list["ListItem"]
    actions: NotRequired[list["Action"]]


class Detail(TypedDict):
    type: Literal["detail"]
    title: NotRequired[str]
    language: NotRequired[str]
    text: NotRequired[str]
    commands: NotRequired[str]
    actions: NotRequired[list["Action"]]


Page = List | Detail


class ListItem(TypedDict):
    title: str
    subtitle: NotRequired[str]
    accessories: NotRequired[list[str]]
    alias: NotRequired[str]
    detail: NotRequired["ListDetail"]
    actions: NotRequired[list["Action"]]


class ListDetail(TypedDict):
    Text: NotRequired[str]
    Commands: NotRequired[list[str]]


class RunAction(TypedDict):
    type: Literal["run"]
    command: str
    title: NotRequired[str]
    shortcut: NotRequired[str]
    inputs: NotRequired[list["FormInput"]]


class EditAction(TypedDict):
    type: Literal["edit"]
    path: str
    title: NotRequired[str]
    shortcut: NotRequired[str]


class ReadAction(TypedDict):
    type: Literal["read"]
    path: str
    title: NotRequired[str]
    shortcut: NotRequired[str]


class OpenAction(TypedDict):
    type: Literal["open"]
    path: NotRequired[str]
    url: NotRequired[str]
    title: NotRequired[str]
    shortcut: NotRequired[str]


class CopyAction(TypedDict):
    type: Literal["copy"]
    text: str
    title: NotRequired[str]
    shortcut: NotRequired[str]


Action = RunAction | EditAction | ReadAction | OpenAction | CopyAction


class TextField(TypedDict):
    type: Literal["textfield"]
    placeholder: str
    title: NotRequired[str]
    default: str


class TextArea(TypedDict):
    type: Literal["textarea"]
    placeholder: str
    title: NotRequired[str]
    default: str


class Dropdown(TypedDict):
    type: Literal["dropdown"]
    placeholder: str
    title: NotRequired[str]
    default: str
    choices: list["DropdownChoice"]


class DropdownChoice(TypedDict):
    title: str
    value: str


FormInput = TextField


hi: List = {
    "type": "list",
    "items": [{"title": "Hello", "actions": []}],
}

json.dump(hi, sys.stdout, indent=2)
