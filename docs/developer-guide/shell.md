# Shell Extensions

## Writing the manifest

When the script is called without arguments, it must return a json manifest describing the extension and its commands.

We will use the `sunbeam query` command to generate this manifest.
The `query` command allows you to manipulate json using the jq syntax.

> ℹ️ Sunbeam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies. An user running a sunbeam extension does not necessarily have jq installed on their system, but he probably has sunbeam installed 😛.

If you are not familiar with jq, I recommend reading this tutorial: <https://earthly.dev/blog/jq-select/>

Note that the script must be executable, and must have a shebang line at the top indicating the interpreter to use.

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
                mode: "list"
            }
        ]
    }'
    exit 0
fi

echo "Not implemented" >&2
exit 1
```

Here we define a single command named `search-docsets`, with a mode set to `list`.

We can run our script to see the generated manifest:

```console
$ ./sunbeam-devdocs
{
  "commands": [
    {
      "mode": "list",
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

Here the command name is `search-docsets`, and the command mode is `list`, so the script must return a view when called with this argument.

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
                mode: "list"
            }
        ]
    }'
    exit 0
fi

# extract the command name from the payload passed as first argument to the script
COMMAND=$(echo "$1" | jq -r '.command')

# When the command name is "search-docsets", the list of docsets is returned
if [ "$COMMAND" = "search-docsets" ]; then
  sunbeam fetch https://devdocs.io/docs/docs.json | sunbeam query '{
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

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "list"
            },
            {
                name: "search-entries",
                title: "Search entries",
                mode: "list",
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

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "search-docsets" ]; then
    # ...
elif [ "$COMMAND" = "search-entries" ]; then
    # we extract the slug param from the payload
    DOCSET=$(echo "$1" | sunbeam query -r '.params.docset')

    sunbeam fetch "https://devdocs.io/docs/$DOCSET/index.json" | sunbeam query --arg docset="$DOCSET" '.entries | {

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

Here we add a new command named `search-entries`, with a `list` mode. We also add a `docset` parameter, which is required.

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
sunbeam run ./sunbeam-devdocs search-entries --docset=go
```

If you run the `search-entries` command without providing the `docset` parameter, a form will will be shown for you to fill the missing parameters.

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
