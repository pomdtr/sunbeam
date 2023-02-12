# Managing extensions

This guide is intended to help you get started with Sunbeam.

## Install you first extension

Sunbeam is a command line launcher, it requires extensions to provide the actual functionality.

You can install an extension from a local directory or a git repository.

Let's install the [file-browser](https://github.com/pomdtr/sunbeam-file-browser) extension from github:

```shell
sunbeam extension install file-browser --git https://github.com/pomdtr/sunbeam-file-browser
```

## Run the extension commands

Once the extension is installed, it becomes available trough the `sunbeam` command.

If you run `sunbeam`, you will see multiple new items available in the root view:

You can also run `sunbeam file-browser` to only see the items provided by the `file-browser` extension.

If you want to go the file browser direcly, select a root item and press `Tab` to show secondary actions,
then select `Copy as Shell Command`.

You can the paste the command in your terminal to go to the file browser.

Feel free to create an alias for the command in you shell configuration file !

```shell
alias downloads="sunbeam file-browser browse --root ~/Downloads"
```

## Upgrading extensions

You can upgrade an extension with the `sunbeam extension upgrade` command.

```shell
sunbeam extension upgrade file-browser
```

If you want to upgrade all extensions, you can use the `--all` flag.

```shell
sunbeam extension upgrade --all
```

## Removing an extension

You can remove an extension with the `sunbeam extension remove` command.

```shell
sunbeam extension remove file-browser
```
