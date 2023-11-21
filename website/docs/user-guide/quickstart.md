# Using Sunbeam

## Installing Extensions

A fresh install of sunbeam is quite boring. In order to make it useful, you need to add some extensions.

Sunbeam extensions are just scripts. In order to install an extension, you only need to pass it's path or url to the `sunbeam extension install` command.

```sh
# Install the devdocs extension from sunbeam repository
sunbeam extension install https://github.com/pomdtr/sunbeam/tree/main/extensions/devdocs.sh
```

> ⚠️ Extensions are not verified, nor sandboxed. They can do anything you can do on your computer. Make sure you trust the source / read the code before installing an extension.

## Upgrading Extensions

Use the `sunbeam extension upgrade --all` command to upgrade all your extensions. `sunbeam extension upgrade <extension>` will upgrade a specific extension.

## Other Extension Commands

- `sunbeam extension list` -> list all installed extensions
- `sunbeam extension remove <extension>` -> uninstall an extension
- `sunbeam extension publish <path>` -> publish an extension as a gist (requires the `SUBEAM_GITHUB_TOKEN` environment variable to be set)
- `sunbeam extension configure <extension>` -> configure an extension preferences (if it has any)

## Running Commands

If you run `sunbeam` without any arguments, it will open the default view, which is a list of all the available commands.

Run a command directly using `sunbeam <extension> [command]`.\
For example, `sunbeam devdocs list-docsets` will show a list of all the available docsets.

Pass parameters to the command using `sunbeam <extension> [command] --param1 value1 --param2 value2`. \
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
