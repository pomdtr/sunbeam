# User Guide

## Managing Extensions

A fresh install of sunbeam is pretty boring. In order to make it useful, you need to add some extensions.

You can manage your extensions using the `sunbeam extension` command.
To get a list of available extensions, run `sunbeam extension browse` command.
It will open the [extension registry](https://github.com/topics/sunbeam-extension) in your browser.

Once you find an extension you like, you can install it using the `sunbeam extension install <url>` command.

> ⚠️ Extensions are not verified, nor sandboxed. They can do anything you can do on your computer. Make sure you trust the source before installing an extension.

For example, let's install the Devdocs extension:

```sh
sunbeam extension install https://github.com/pomdtr/sunbeam-devdocs
```

Now, you can run `sunbeam devdocs` to open the devdocs extension, and search for documentation in your terminal.

Use `enter` to select a docset, then type to search for a specific entry, and press `enter` again to open it in your browser.
Or, if you just want to copy the entry url, press `tab` to show the list of actions, and select `copy url`.

If you regularly use a specific docset, you can skip the docset selection view by using `sunbeam devdocs list-entries --slug go` for example.
Use `sunbeam devdocs --help` to get a list of available commands.

Other managements commands are available:

- sunbeam extension list: list installed extensions
- sunbeam extension remove: uninstall an extension
    - ex: `sunbeam extension remove devdocs`
- sunbeam extension rename: rename an extension
    - ex: `sunbeam extension rename devdocs dd`
- sunbeam extension upgrade: upgrade an extension
    - ex: `sunbeam extension upgrade devdocs` or `sunbeam extension upgrade --all`

## Root Commands

By default, if you run sunbeam without any arguments, it will show an usage message.

```
$ sunbeam
Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.

Usage:
  sunbeam [flags]
  sunbeam [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  extension          Manage extensions
  fetch              Simple http client inspired by curl
  help               Help about any command
  query              Transform or generate JSON using a jq query
  run                Run an extension without installing it
  validate           Validate a Sunbeam schema

Flags:
  -h, --help      help for sunbeam
  -v, --version   version for sunbeam

Use "sunbeam [command] --help" for more information about a command.
```

As soon as you install an extension, a list of commands will be shown instead.
