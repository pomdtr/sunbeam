import sys

if sys.version_info < (3, 11):
    from typing_extensions import Literal, NotRequired, TypedDict, Union
else:
    from typing import Literal, NotRequired, TypedDict, Union


class ActionFetchUrl(TypedDict):
    type: Literal["fetch-url"]
    title: NotRequired[str]
    key: NotRequired[str]
    url: str
    method: NotRequired[Literal["GET", "POST", "PUT", "DELETE"]]
    body: NotRequired[str]
    headers: NotRequired[dict[str, str]]
    inputs: NotRequired[list["Input"]]


class ActionCopyText(TypedDict):
    type: Literal["copy-text"]
    title: NotRequired[str]
    text: str
    key: NotRequired[str]


class ActionOpenFile(TypedDict):
    type: Literal["open-file"]
    title: NotRequired[str]
    key: NotRequired[str]
    path: NotRequired[str]


class ActionOpenUrl(TypedDict):
    type: Literal["open-url"]
    title: NotRequired[str]
    key: NotRequired[str]
    url: NotRequired[str]


class ActionRunCommand(TypedDict):
    type: Literal["run-command"]
    title: NotRequired[str]
    key: NotRequired[str]
    command: str
    dir: NotRequired[str]
    onSuccess: NotRequired[Literal["reload", "exit", "push"]]
    inputs: NotRequired[list["Input"]]


class ActionReadFile(TypedDict):
    type: Literal["read-file"]
    title: NotRequired[str]
    key: NotRequired[str]
    path: str


Action = Union[
    ActionFetchUrl,
    ActionCopyText,
    ActionOpenFile,
    ActionOpenUrl,
    ActionRunCommand,
    ActionReadFile,
]


class InputTextField(TypedDict):
    name: str
    title: str
    type: Literal["textfield"]
    placeholder: NotRequired[str]
    default: NotRequired[str]
    secure: NotRequired[bool]


class InputCheckbox(TypedDict):
    name: str
    title: str
    type: Literal["checkbox"]
    default: NotRequired[bool]
    label: NotRequired[str]
    trueSubstitution: NotRequired[str]
    falseSubstitution: NotRequired[str]


class InputTextArea(TypedDict):
    name: str
    title: str
    type: Literal["textarea"]
    placeholder: NotRequired[str]
    default: NotRequired[str]


class InputDropdown(TypedDict):
    name: str
    title: str
    type: Literal["dropdown"]
    items: list[dict[str, str]]
    default: NotRequired[str]


Input = Union[
    InputTextField,
    InputCheckbox,
    InputTextArea,
    InputDropdown,
]

class StaticPreview(TypedDict):
    text: str
    language: NotRequired[str]

class DynamicPreview(TypedDict):
    command: str
    dir: NotRequired[str]
    language: NotRequired[str]

Preview = Union[StaticPreview, DynamicPreview]


class List(TypedDict):
    type: Literal["list"]
    title: NotRequired[str]
    emptyText: NotRequired[str]
    showPreview: NotRequired[bool]
    actions: NotRequired[list[Action]]
    items: list["ListItem"]


class ListItem(TypedDict):
    title: str
    id: NotRequired[str]
    subtitle: NotRequired[str]
    preview: NotRequired[Preview]
    accessories: NotRequired[list[str]]
    actions: NotRequired[list[Action]]


class Detail(TypedDict):
    type: Literal["detail"]
    title: NotRequired[str]
    preview: NotRequired[Preview]
    actions: NotRequired[list[Action]]


class Page(TypedDict):
    List: NotRequired[List]
    Detail: NotRequired[Detail]
