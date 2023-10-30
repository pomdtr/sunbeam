# Command

## Copy

Copy text to the clipboard.

```json
{
    // the title of the action (required)
    "title": "Copy",
    // the key to trigger the action (optional)
    "key": "c",
    // the type of the action (required)
    "type": "copy",
    // the text to copy (required)
    "text": "hello world",
    // whether to exit sunbeam after copying the text (optional)
    // if not specified, sunbeam will not exit
    "exit": true
}
```

## Open

Open a url or a file with the default app or a specific app.

```json
{

    // the title of the action (required)
    "title": "Open",
    // the key to trigger the action (optional)
    "key": "o",
    // the title of the action (required)
    "type": "open",
    // the target to open (required)
    // target can be a url or a path to a file
    "target": "https://pomdtr.github.io/sunbeam",
    // the app to use to open the target (optional)
    // if not specified, the default app will be used
    "app": {
        "windows": "chrome",
        "macos": "google chrome",
        "linux": "google-chrome"
    },
    // whether to exit sunbeam after opening the target (optional)
    // if not specified, sunbeam will not exit
    "exit": true
}
```

## Run

Run a custom command defined in the extension manifest.

```json
{
    // the title of the action (required)
    "title": "View Readme",
    // the key to trigger the action (optional)
    "key": "v",
    // the type of the action (required)
    "type": "run",
    // the command to run (must be defined in the extension manifest) (required)
    "command": "view-readme",
    // the arguments to pass to the command (optional)
    "params": {
        "full_name": "pomdtr/sunbeam"
    },
}
```

## Reload

Reload the current view.

```json
{
    // the title of the action (required)
    "title": "Reload",
    // the key to trigger the action (optional)
    "key": "r",
    // the type of the action (required)
    "type": "reload",
    // override the current params (optional)
    "params": {
        "full_name": "pomdtr/sunbeam" // key must match the name of the param
    }
}
```

## Exit

Exit sunbeam.

```json
{
    // the title of the action (required)
    "title": "Exit",
    // the key to trigger the action (optional)
    "key": "q",
    // the type of the action (required)
    "type": "exit"
}
```
