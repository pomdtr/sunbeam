from typing import TypedDict, NotRequired, Literal


class Param(TypedDict):
    default: NotRequired[str | int | bool]


class RunAction(TypedDict):
    command: str
    params: dict[str, str | int | bool | Param]


class ReloadAction(TypedDict):
    params: dict[str, str | int | bool | Param]


class OpenAction(TypedDict):
    url: NotRequired[str]
    path: NotRequired[str]


class CopyAction(TypedDict):
    text: str
    exit: NotRequired[bool]


class EditAction(TypedDict):
    path: str
    reload: NotRequired[bool]


ActionType = Literal["copy", "open", "run", "reload", "exit", "edit", "exit"]


class Action(TypedDict):
    title: str
    key: NotRequired[str]
    type: ActionType
    copy: NotRequired[CopyAction]
    reload: NotRequired[ReloadAction]
    open: NotRequired[OpenAction]
    run: NotRequired[RunAction]
    edit: NotRequired[EditAction]


class ListItemDetail(TypedDict):
    markdown: NotRequired[str]
    text: NotRequired[str]


class ListItem(TypedDict):
    title: str
    id: NotRequired[str]
    subtitle: NotRequired[str]
    detail: NotRequired[ListItemDetail]
    accessories: NotRequired[list[str]]
    actions: list[Action]


class List(TypedDict):
    items: NotRequired[list[ListItem]]
    showDetail: NotRequired[bool]
    actions: NotRequired[list[Action]]
    emptyText: NotRequired[str]


class Detail(TypedDict):
    markdown: NotRequired[str]
    text: NotRequired[str]
    actions: NotRequired[list[Action]]
