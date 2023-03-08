# Reference

## List

```javascript
{
  "type": "list", // required
  "showPreview": false, // optional, default: false
  "generateItems": false, // optional, default: false
  "items": [] // required, see ListItem
}
```

## Detail

```javascript
{
  "type": "detail",
  "content": "preview text", // required, see Preview
  "actions": [...] // optional, see Action
}
```

## ListItem

```javascript
{
  "title": "Item title", // required
  "subtitle": "Item subtitle", // optional
  "accessories": ["Accessory 1", "Accessory 2"], // optional
  "actions": [] // optional, see Action
}
```

## Action

```javascript
{
  "type": "open", // required. See Action types
  "title": "Action title",
  "shortcut": "ctrl+o" // optional
}
```

### open

Open an URI in the default application/browser.

```javascript
{
  "type": "open", // required
  "title": "Open in browser", // optional, default: "Open"
  "target": "https://example.com" // required
}
```

### copy

Copy text to the system clipboard.

```javascript
{
  "type": "copy", // required
  "title": "Copy to Clipboard", // optional, default: "Copy"
  "text": "username" // required
}
```

### run

Run a command, and display output on stdout.

```javascript
{
  "type": "run", // required
  "title": "Run", // required
  "command": "printf", // required
  "args": ["Hello World"] // optional
}
```

### push

Push a new page to the navigation stack

```javascript
{
  "type": "push", // required
  "title": "Push", // required
  "command": "my-command", // required
  "args": ["list-items"] // optional
}
```

### reload

Reload the current page

```javascript
{
  "type": "reload", // required
  "title": "Reload" // optional, default: "Reload"
}
```

## Preview

The preview can be a string or a command. If it's a command, the output will refreshed every time the user changes the selection.

```javascript
{
  "text": "preview text"
}
```

```javascript
{
  "command": "my-command",
  "args": ["preview"]
}
```
