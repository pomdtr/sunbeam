## Page

**POSSIBLE VALUES**

- [List](#list)
- [Detail](#detail)

## Command

**POSSIBLE VALUES**

- string
- &lt;string, string&gt;
- object
  - `name`: string
  - `args`: string[]
  - `input`: string
  - `dir`: string

## Action

**POSSIBLE VALUES**

- object
  - `type`: `'copy'` - The type of the action.
  - `title`: string - The title of the action.
  - `text`: string - The text to copy.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `key`: string - The key used as a shortcut.
- object
  - `type`: `'open'` - The type of the action.
  - `title`: string - The title of the action.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `key`: string - The key used as a shortcut.
  - `target`: string - The target to open.
- object
  - `type`: `'exit'` - The type of the action.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `title`: string - The title of the action.
  - `key`: string - The key used as a shortcut.
- object
  - `type`: `'reload'` - The type of the action.
  - `title`: string - The title of the action.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `key`: string - The key used as a shortcut.
  - `command`: [Command](#command)
- object
  - `type`: `'run'` - The type of the action.
  - `title`: string - The title of the action.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `key`: string - The key used as a shortcut.
  - `command`: [Command](#command)
  - `reloadOnSuccess`: boolean - Whether to reload the page when the command succeeds.
- object
  - `type`: `'push'` - The type of the action.
  - `title`: string - The title of the action.
  - `key`: string - The key used as a shortcut.
  - `inputs`: [Input](#input)[] - The inputs to show when the action is run.
  - `page`: [TextOrCommand](#textorcommand)

## Input

**POSSIBLE VALUES**

- object
  - `name`: string - The name of the input.
  - `title`: string - The title of the input.
  - `type`: `'textfield'` - The type of the input.
  - `placeholder`: string - The placeholder of the input.
  - `optional`: boolean - Whether the input is optional.
  - `default`: string - The default value of the input.
  - `secure`: boolean - Whether the input should be secure.
- object
  - `name`: string - The name of the input.
  - `title`: string - The title of the input.
  - `optional`: boolean - Whether the input is optional.
  - `type`: `'checkbox'` - The type of the input.
  - `default`: boolean - The default value of the input.
  - `label`: string - The label of the input.
  - `trueSubstitution`: string - The text substitution to use when the input is true.
  - `falseSubstitution`: string - The text substitution to use when the input is false.
- object
  - `name`: string - The name of the input.
  - `title`: string - The title of the input.
  - `type`: `'textarea'` - The type of the input.
  - `optional`: boolean - Whether the input is optional.
  - `placeholder`: string - The placeholder of the input.
  - `default`: string - The default value of the input.
- object
  - `name`: string - The name of the input.
  - `title`: string - The title of the input.
  - `optional`: boolean - Whether the input is optional.
  - `type`: `'dropdown'` - The type of the input.
  - `items`: object[] - The items of the input.
- `title`: string - The title of the item.
- `value`: string - The value of the item.
  - `default`: string - The default value of the input.

## TextOrCommand

**POSSIBLE VALUES**

- string
- object
  - `command`: [Command](#command)
- object
  - `text`: string

## List

**PROPERTIES**

- `type`: `'list'` - The type of the response.
- `title`: string - The title of the page.
- `onQueryChange`: [Command](#command)
- `emptyView`: object
  - `text`: string - The text to show when the list is empty.
  - `actions`: [Action](#action)[] - The actions to show when the list is empty.
- `showPreview`: boolean - Whether to show the preview on the right side of the list.
- `items`: [Listitem](#listitem)[]

## Listitem

**PROPERTIES**

- `title`: string - The title of the item.
- `id`: string - The id of the item.
- `subtitle`: string - The subtitle of the item.
- `preview`: [TextOrCommand](#textorcommand)
- `accessories`: string[] - The accessories to show on the right side of the item.
- `actions`: [Action](#action)[] - The actions attached to the item.

## Detail

A detail view displayign a preview and actions.

**PROPERTIES**

- `type`: `'detail'` - The type of the response.
- `title`: string - The title of the page.
- `preview`: [TextOrCommand](#textorcommand)
- `actions`: [Action](#action)[] - The actions attached to the detail view.