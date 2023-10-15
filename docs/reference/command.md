# Command

Command can be attached to actions array in a view, or returned from a script extension when the mode is `no-view` or `tty`.

## Copy

Copy text to the clipboard.

```json
{
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

Run a command defined in the manifest of the current extension.

```json
{
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
    "type": "reload",
    // override the current params (optional)
    "params": {
        "show-hidden": true
    }
}
```

## Exit

Exit sunbeam.

```json
{
    "type": "exit"
}
```

## Pop

Pop the current view from the navigation stack.

```json
{
    "type": "pop",
    // wether the view should be reloaded after popping (optional)
    "reload": true
}
```
