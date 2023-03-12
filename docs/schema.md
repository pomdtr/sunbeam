# Reference

## List

```javascript
{
  "type": "list", // required
  "showDetail": false, // optional, default: false
  "generateItems": false, // optional, default: false
  "items": [
    {"title": "Item 1"},
    {"title": "Item 2"}
  ] // required, see ListItem
}
```

## Detail

```javascript
{
  "type": "detail", // required
  "text": "detail text",
  "command": "print-detail-cmd",
  "actions": [
    {"type": "open", "target": "https://example.com"},
    {"type": "copy", "text": "username"}
  ] // optional, see Action
}
```

## ListItem

```javascript
{
  "title": "Item title", // required
  "subtitle": "Item subtitle", // optional
  "accessories": ["Accessory 1", "Accessory 2"], // optional
  "actions": [
    { "type": "open", "target": "https://example.com" },
    { "type": "copy", "text": "username" }
  ] // optional, see Action
}
```

## Actions

### open

Open an URI in the default application/browser.

```javascript
{
  "type": "open", // required
  "title": "Open in browser", // optional, default: "Open"
  "shortcut": "ctrl+o", // optional
  "target": "https://example.com" // required
}
```

### copy

Copy text to the system clipboard.

```javascript
{
  "type": "copy", // required
  "shortcut": "ctrl+y", // optional
  "title": "Copy to Clipboard", // optional, default: "Copy"
  "text": "username" // required
}
```

### run

Run a command, and handle the output depending on the `onSuccess` property.

```javascript
{
  "type": "run", // required
  "title": "Run", // required
  "onSuccess": "push", // optional, default: null
  "command": ["my-command"], // required
}
```

### push

Push a new page to the navigation stack

```javascript
{
  "type": "read", // required
  "title": "Read", // required
  "path": "path/to/page.json" // required
}
```

## FormInput

## textfield

```javascript
{
  "type": "textfield", // required
  "title": "Username", // required
  "name": "username", // required
}
```

## textarea

```javascript
{
  "type": "textarea", // required
  "title": "Description", // required
  "name": "description", // required
}
```

## dropdown

```javascript
{
  "type": "dropdown", // required
  "title": "Language", // required
  "name": "language", // required
  "options": [
    {"title": "English", "value": "en"},
    {"title": "French", "value": "fr"}
  ] // required
}
```
