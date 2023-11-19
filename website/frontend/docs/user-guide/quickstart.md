# Using Sunbeam

## Installing Extensions

A fresh install of sunbeam is quite boring. In order to make it useful, you need to add some extensions.

Sunbeam extensions are just scripts. In order to install an extension, you only need to add it's path or url to the config file.

```json
{
    "extensions": {
        "tldr": "~/sunbeam/tldr.sh",
        "devdocs": "https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/devdocs.sh",
    }
}
```

> ⚠️ Extensions are not verified, nor sandboxed. They can do anything you can do on your computer. Make sure you trust the source before installing an extension.

## Configuring Extensions

Some extensions require additional configuration. You can configure an extension by adding a `preferences` dict to the config file.

```json
{
    "extensions": {
        "github": {
            "origin": "~/sunbeam/github.sh",
            "preferences": {
                "token": "xxxx"
            }
        }
    },
}
```

## Upgrading Extensions

Use the `sunbeam upgrade` command to upgrade all your extensions. `sunbeam upgrade <extension>` will upgrade a specific extension.

## Running Commands

If you run `sunbeam` without any arguments, it will open the default view, which is a list of all the available commands.

You can also run a command directly using `sunbeam <extension> [command]`.\
For example, `sunbeam devdocs list-docsets` will show a list of all the available docsets.

You can also pass parameters to the command using `sunbeam <extension> [command] --param1 value1 --param2 value2`. \
For example, `sunbeam devdocs list-entries --docset go` will list all the entries in the go docset.


## Using the Sunbeam UI

Sunbeam is designed to be used with your keyboard. Depending on the current view, multiple keyboard shortcuts are available:

- all views:
    - `ctrl+r` -> refresh the current view
    - `ctrl+c` -> exit sunbeam
    - `escape` -> go back to the previous page
- root view:
    - `ctrl+e` -> edit sunbeam config
    - `alt+enter` -> run query as a shell command
- list view:
    - `up` / `ctrl+k` -> move selection up
    - `down` / `ctrl+j` -> move selection down
    - `enter` -> execute the selected command
    - `tab` -> show the available actions for the selected item
- detail view:
    - `up`, `k` -> scroll one line up
    - `down`, `j` -> scroll one line down
    - `ctrl+u` -> scroll half a page up
    - `ctrl+d` -> scroll half a page down
    - `q` -> exit sunbeam
    - `tab` -> show the available actions
