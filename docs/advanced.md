# Advanced

## Displaying Text

If you want to display a block of text to the user, you can use the `detail` page type:

```json
{
  "type": "detail",
  "preview": {
    "text": "Hello World!"
  }
}
```

Detail view support actions, so you can add a button to the page:

```json
{
  "type": "detail",
  "preview": {
    "text": "Hello World!"
  },
  "actions": [
    {
      "type": "run",
      "command": "echo Hello World!"
    }
  ]
}
```

## Handling User Input

### Adding Inputs to Actions

You may want to ask the user for some input before running a command.

You can add an array of inputs to every sunbeam actions, and reference their value in the command.

```json
{
  "type": "run",
  "command": "echo ${input:name}",
  "inputs": [
    {
      "name": "name",
      "type": "textfield",
      "title": "Name",
      "title": "What's your name?"
    }
  ]
}
```

When activating the action, the user will be prompted for the input value.
Sunbeam support the following input types:

- `textfield`: A simple text input
- `textarea`: A multiline text input
- `checkbox`: A checkbox
- `dropdown`: A dropdown menu

### Showing a Form

Alternatively, you can use a Form Page. Most of the time, adding inputs to actions is enough, but if you want to display a form as your first page, the form page is the way to go.

```json
{
  "type": "form",
  "submitAction": {
    "type": "run",
    "command": "echo ${input:name}",
    "inputs": [
      {
        "name": "name",
        "type": "textfield",
        "title": "Name"
      }
    ]
  }
}
```
