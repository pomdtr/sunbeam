"""
Types for sunbeam scripts
"""

__version__ = "0.1.0"

from dataclasses import dataclass, asdict, field
from typing import Literal
import json


@dataclass
class ListItemDetail:
    text: str = ""
    command: str = ""
    dir: str = ""
    language: str = ""


@dataclass
class Page:
    def json(self):
        def dict_factory(fields):
            return {field[0]: field[1] for field in fields if field[1]}

        return json.dumps(asdict(self, dict_factory=dict_factory))


@dataclass
class List(Page):
    items: list["ListItem"]
    title: str = ""


@dataclass
class ListItem:
    title: str
    id: str = ""
    subtitle: str = ""
    detail: ListItemDetail = None
    actions: list["Action"] = field(default_factory=list)


@dataclass
class Detail(Page):
    actions: list["Action"]
    text: str = ""
    command: str = ""
    dir: str = ""
    language: str = ""
    inputs: list["Input"] = field(default_factory=list)


@dataclass
class OpenAction:
    url: str
    path: str
    type: str = field(init=False)
    inputs: list["Input"] = field(default_factory=list)

    def __post_init__(self):
        self.type = "open"


@dataclass
class CopyAction:
    text: str
    inputs: list["Input"] = field(default_factory=list)
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "copy"


@dataclass
class ReadAction:
    path: str
    inputs: list["Input"] = field(default_factory=list)
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "read"


@dataclass
class FetchAction:
    url: str
    body: str
    method: str
    headers: dict[str, str] = field(default_factory=dict)
    inputs: list["Input"] = field(default_factory=list)
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "fetch"


@dataclass
class RunAction:
    command: str
    onSuccess: Literal["push", "pop", "exit"]
    dir: str = ""
    inputs: list["Input"] = field(default_factory=list)
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "run"


Action = RunAction | ReadAction | FetchAction | CopyAction | OpenAction


@dataclass
class TextField:
    text: str
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "textfield"


@dataclass
class TextArea:
    text: str
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "textarea"


@dataclass
class DropDown:
    text: str
    options: list[str]
    type: str = field(init=False)

    def __post_init__(self):
        self.type = "dropdown"


Input = TextField | TextArea | DropDown

page: Page = List(
    items=[
        ListItem(
            title="foo",
            actions=[
                RunAction(
                    command="echo foo",
                    onSuccess="push",
                ),
            ],
        )
    ]
)
