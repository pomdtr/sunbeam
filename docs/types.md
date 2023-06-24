## List(list)

**PARAMETERS**

- `list`: [List](#list)

**RETURNS**

[List](#list)

## Detail(detail)

**PARAMETERS**

- `detail`: [Detail](#detail)

**RETURNS**

[Detail](#detail)

## Item(item)

**PARAMETERS**

- `item`: [Listitem](#listitem)

**RETURNS**

[Listitem](#listitem)

## Action(action)

**PARAMETERS**

- `action`: [Action](#action)

**RETURNS**

[Action](#action)

## Input(input)

**PARAMETERS**

- `input`: [Input](#input)

**RETURNS**

[Input](#input)

## Page

**POSSIBLE VALUES**

- [List](#list)
- [Detail](#detail)
- [Form](#form)

## Command

**POSSIBLE VALUES**

- string
- &lt;string, string&gt;
- object
  - `name`: string
  - `args`: string[]
  - `input`: string
  - `dir`: string

## OnSuccess

**POSSIBLE VALUES**

- `'copy'`
- `'paste'`
- `'open'`
- `'reload'`

## Request

**POSSIBLE VALUES**

- string
- object
  - `url`: string - The URL to request.
  - `method`: string - The HTTP method to use.
  - `headers`: object - The headers to send.
    - `__index`: any
  - `body`: string - The body to send.

## Listitem

**PROPERTIES**

- `title`: string - The title of the item.
- `id`: string - The id of the item.
- `subtitle`: string - The subtitle of the item.
- `preview`: [Preview](#preview)
- `accessories`: string[] - The accessories to show on the right side of the item.
- `actions`: [Action](#action)[] - The actions attached to the item.

## Preview

**PROPERTIES**

- `command`: [Command](#command)
- `request`: [Request](#request)
- `text`: string
- `file`: string
- `expression`: string
- `__index`: any

## Form

**PROPERTIES**

- `type`: `'form'` - The type of the response.
- `title`: string - The title of the page.
- `submitAction`: [Action](#action)