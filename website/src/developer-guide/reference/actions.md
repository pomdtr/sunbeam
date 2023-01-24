# Actions

## Copy to Clipboard

```jsonc
{
    "type": "copy-to-clipboard", // required
    "title": "Copy to clipboard", // optional, defaults to "Copy to Clipboard"
    "shortcut": "ctrl+c", // optional
    "text": "Hello World" // required
}
```

## Open in Browser

```jsonc
{
    "type": "open-url", // required
    "title": "Open Google", // optional, defaults to "Open Url"
    "shortcut": "ctrl+g", // optional
    "url": "https://www.google.com" // required
}
```

## Run Command

```jsonc
{
    "type": "run-command", // required
    "title": "Run Command", // required
    "shortcut": "ctrl+r", // optional
    "with": { // Parameters to pass to the command
        "key": "value", // pass a fixed value
        "name": { // show a form input to the user
            "type": "textfield",
            "title": "Name",
        }
    },
    "command": "echo Hello World" // required
}
```

## Reload Page

```jsonc
{
    "type": "reload-page", // required
    "title": "Reload Page", // optional, defaults to "Reload Page"
    "shortcut": "ctrl+r", // optional
    "with": {
        "key": "value", // override a parameter when reloading the page
    }
}
```
