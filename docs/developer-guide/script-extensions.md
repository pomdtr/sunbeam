# Writing your Extension

## Initial setup

To create a sunbeam extension, we only need a single script.

```sh
# Create a script named sunbeam-devdocs and make it executable
touch sunbeam-extension && chmod +x sunbeam-devdocs
```

## Writing the manifest

When the script is called without arguments, it must return a json manifest describing the extension and its commands.

We will use the `sunbeam query` command to generate this manifest.
The `query` command allows you to manipulate json using the jq syntax.

> ℹ️ Sunbeam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies. An user running a sunbeam extension does not necessarily have jq installed on their system, but he probably has sunbeam installed 😛.

If you are not familiar with jq, I recommend reading this tutorial: <https://earthly.dev/blog/jq-select/>

Note that the script must be executable, and must have a shebang line at the top indicating the interpreter to use.
The following tutorial will use `#!/bin/sh`, but you can use any interpreter you want (e.g. `/usr/bin/env python3` or `/usr/bin/env -S deno run -A`).

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "page"
            }
        ]
    }'
    exit 0
fi

# TODO: handle commands
```

Here we define a single command named `search-docsets`, with a `view` mode.

We can run our script to see the generated manifest:

```console
$ ./sunbeam-devdocs
{
  "commands": [
    {
      "mode": "page",
      "name": "search-docsets",
      "title": "Search docsets"
    }
  ],
  "description": "Search the devdocs.io documentation",
  "title": "Devdocs"
}
```

Or use sunbeam to see the generated command:

```console
$ sunbeam run ./sunbeam-devdocs --help
Search the devdocs.io documentation

Usage:
  extension [flags]
  extension [command]

Available Commands:
  search-docsets Search docsets

Flags:
  -h, --help   help for extension

Use "extension [command] --help" for more information about a command.
```

## Handling commands

When the user run a command, the script is called with the command name as first argument.

Here the command name is `search-docsets`, and the command mode is `view`, so the script must return a view when called with this argument.

We will use the `sunbeam fetch` command to fetch the list of docsets from the devdocs api. The `fetch` command allows you to perform http requests (with an api similar to curl).

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "page"
            }
        ]
    }'
    exit 0
fi

# When the command name is "search-docsets", the list of docsets is returned
if [ "$1" = "search-docsets" ]; then
  sunbeam fetch https://devdocs.io/docs/docs.json | sunbeam query '{
    type: "list",
    items: map({
      title: .name,
      subtitle: (.release // "latest"),
      accessories: [ .slug ],
      actions: [
        {
            type: "open",
            title: "Open in Browser",
            target: "https://devdocs.io/\(.slug)",
            exit: true
        }
      ]
    })
  }'
fi
```

Here we pipe the json catalog to the `query` command to transform it into a list, with each docset as an item. We also add an action to open the docset in the user default browser.

Let's run our script to see the generated view:

```console
sunbeam run ./sunbeam-devdocs search-docsets
```

Note that since we only have one command, we can omit the command name:

```console
sunbeam run ./sunbeam-devdocs
```

## Chaining Commands

This extension is still pretty basic. Sunbeam really shines when you start chaining commands together. Let's add a new command to search the documentation of a specific docset.

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "page"
            },
            {
                name: "search-entries",
                title: "Search entries",
                mode: "page",
                params: [
                    {
                        name: "docset",
                        type: "string",
                        required: true
                    }
                ]
            }
        ]
    }'
    exit 0
fi

if [ "$1" = "search-docsets" ]; then
    # ...
elif [ "$1" = "search-entries" ]; then
    # we extract the slug param from stdin
    DOCSET=$(sunbeam query -r '.params.docset')

    sunbeam fetch "https://devdocs.io/docs/$DOCSET/index.json" | sunbeam query --arg docset="$DOCSET" '.entries | {
        type: "list",
        items: map({
            title: .name,
            subtitle: .type,
            actions: [
                {
                    title: "Open in Browser",
                    type: "open",
                    target: "https://devdocs.io/\($docset)/\(.path)",
                    exit: true
                },
                {
                    type: "copy",
                    title: "Copy URL",
                    key: "c",
                    text: "https://devdocs.io/\($docset)/\(.path)",
                    exit: true
                }
            ]
        })
    }'
fi
```

Here we add a new command named `search-entries`, with a `view` mode. We also add a `docset` parameter, which is required.

When the user run this command, the script is called with the command name as first argument, and the parameters as json on stdin.

If we run run `sunbeam run ./sunbeam-devdoc --help` again, we can see the new command:

```console
$ sunbeam run ./sunbeam-devdocs --help
Search the devdocs.io documentation

Usage:
  extension [flags]
  extension [command]

Available Commands:
  search-docsets Search docsets
  search-entries Search entries

Flags:
  -h, --help   help for extension

Use "extension [command] --help" for more information about a command.
```

Let's run the command to see the generated view:

```console
sunbeam run ./sunbeam-devdocs  search-entries --docset=go
```

If we want to be able to go from the docsets list to the entries list, we can add `run` action to the docsets list:

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    ...
fi

if [ "$1" = "search-docsets" ]; then
    sunbeam fetch https://devdocs.io/docs/docs.json | sunbeam query '{
        type: "list",
        items: map({
          title: .name,
          subtitle: (.release // "latest"),
          accessories: [ .slug ],
          actions: [
            {
                type: "run",
                title: "Search \(.name) entries",
                command: "search-entries",
                params: {
                    docset: .slug
                }
            },
            {
                type: "open",
                title: "Open in Browser",
                target: "https://devdocs.io/\(.slug)",
                exit: true
            }
          ]
        })
    }'
elif [ "$1" = "search-entries" ]; then
    # ...
fi
```

Now we can start by listing the docsets, select the one we are interested in, and then search the entries of this docset.

## Installing your extension

Now that we have a working extension, we can install it using the `sunbeam extension install` command.

```console
sunbeam extension install ./sunbeam-devdocs
sunbeam devdocs --help
```

Now we can run the extension from anywhere using the `sunbeam devdocs` command. The `devdocs` alias is based on the name of the script, stripped of the `sunbeam-` prefix if it exists. If you prefer, you can specify a custom alias using the `--alias` flag.

```console
sunbeam extension install ./sunbeam-devdocs --alias=dd
sunbeam dd --help
```

The search-docsets command will also appears in the root list (only commands without required parameters are shown in the root list).

Note that if you update the manifest of the extension, you will need to upgrade it using the `sunbeam extension upgrade` to see the changes.

```console
sunbeam extension upgrade devdocs
```

> ℹ️ The source code of this extension is available here: <https://github.com/pomdtr/sunbeam-devdocs/blob/main/sunbeam-extension>. Use `sunbeam extension install https://raw.githubusercontent.com/pomdtr/sunbeam-extensions/main/extensions/devdocs.sh` to install it.

## Publishing an extension

Just create a git repository containing your extension, and push it to github. User will be able to install it using:

```console
sunbeam extension install <raw-url>
```

## Make your extension discoverable

Add the `sunbeam` topic to your github repository to make it discoverable to all sunbeam users.

Make sure to add a `README.md` file to your repository, so that users can learn more about your extensions before installing them.

## What's next?

This guide is only a quick introduction to sunbeam extensions. There are many more features available, such as:
  - Gathering user input using `form` view
  - Displaying markdown using the `detail` view
  - Dynamically generate items using the `reload` property of the `list` view
  - Interact with remote extension using http requests
