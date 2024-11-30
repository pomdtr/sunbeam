# Devdocs

## Writing the manifest

When the script is called without arguments, it must return a json manifest describing the extension and its commands.

We will use the `jq` to generate this manifest. jq is a command-line JSON processor. It is available on most linux distributions, and can be installed on macOS using homebrew.

If you are not familiar with jq, I recommend reading this tutorial: <https://earthly.dev/blog/jq-select/>

Note that the script must be executable, and must have a shebang line at the top indicating the interpreter to use.

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    jq -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "filter"
            }
        ]
    }'
    exit 0
fi

echo "Not implemented" >&2
exit 1
```

Here we define a single command named `search-docsets`, with a mode set to `filter`.

We can run our script to see the generated manifest:

```console
$ ./devdocs.sh
{
  "commands": [
    {
      "mode": "filter",
      "name": "search-docsets",
      "title": "Search docsets"
    }
  ],
  "description": "Search the devdocs.io documentation",
  "title": "Devdocs"
}
```

And use the `sunbeam validate` command to validate the manifest:

```console
$ ./devdocs.sh | sunbeam validate manifest
✅ Manifest is valid!
```

Now that we have a valid manifest, we can install the extension.

```console
sunbeam extension install ./devdocs.sh
```

We can now run `sunbeam devdocs --help` to see the generated help, and `sunbeam devdocs` to list the root commands (defined in the `root` field of the manifest).

## Handling commands

When the user run a command, the script is called with the command name as first argument. Let's implement the `search-docsets` command.

The `search-docsets` command has a `filter` mode, so the script must return a [valid list](./../../reference/schemas/list.md) when called with this argument.

```sh
#!/bin/sh

set -eu

# When the number of arguments is 0, return the manifest
if [ $# -eq 0 ]; then
    jq -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "filter"
            }
        ]
    }'
    exit 0
fi

# extract the command name from the payload passed as first argument to the script
COMMAND=$(echo "$1" | jq -r '.command')

# When the command name is "search-docsets", the list of docsets is returned
if [ "$COMMAND" = "search-docsets" ]; then
  curl https://devdocs.io/docs/docs.json | jq '{
    items: map({
      title: .name,
      subtitle: (.release // "latest"),
      accessories: [ .slug ],
      actions: [
        {
            title: "Open in Browser",
            type: "open",
            url: "https://devdocs.io/\(.slug)",
        }
      ]
    })
  }'
fi
```

Here we pipe the json catalog to the `query` command to transform it into a list, with each docset as an item. We also add an action to open the docset in the user default browser.

Let's run our script to see the generated view:

```console
sunbeam devdocs search-docsets
```

## Chaining Commands

This extension is still pretty basic. Sunbeam really shines when you start chaining commands together. Let's add a new command to search the documentation of a specific docset.

```sh
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Devdocs",
        description: "Search the devdocs.io documentation",
        commands: [
            {
                name: "search-docsets",
                title: "Search docsets",
                mode: "filter"
            },
            {
                name: "search-entries",
                title: "Search entries",
                mode: "filter",
                params: [
                    {
                        name: "docset",
                        type: "string",
                        title: "Docset Slug"
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
    DOCSET=$(echo "$1" | jq -r '.params.docset')

    curl "https://devdocs.io/docs/$DOCSET/index.json" | jq --arg docset "$DOCSET" '.entries | {

        items: map({
            title: .name,
            subtitle: .type,
            actions: [
                {
                    title: "Open in Browser",
                    type: "open",
                    url: "https://devdocs.io/\($docset)/\(.path)",
                },
                {
                    type: "copy",
                    title: "Copy URL",
                    key: "c",
                    text: "https://devdocs.io/\($docset)/\(.path)",
                }
            ]
        })
    }'
fi
```

Here we add a new command named `search-entries`, with a `filter` mode. We also add a `docset` parameter, which is required.

If we run `sunbeam devdocs --help` again, we can see the new command help:

```console
$ sunbeam devdocs --help
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
sunbeam devdocs search-entries --docset=go
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
    curl https://devdocs.io/docs/docs.json | jq '{
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
                url: "https://devdocs.io/\(.slug)",
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

## Adding new root items

When we installed the extension using `sunbeam extension install ./devdocs.sh`, an entry was added to the `extensions` map in `~/.config/sunbeam/sunbeam.json`.

```json
{
  "extensions": {
    "devdocs": {
      "origin": "<path-to-extension>/devdocs.sh"
    }
  }
}
```

As a user of the extension, we can add shortcuts to specific docsets:

```json
{
  "extensions": {
    "devdocs": {
      "origin": "<path-to-extension>/devdocs.sh",
      "root": [
        {
          "title": "Search Go documentation",
          "command": "search-entries",
          "params": {
            "docset": "go"
          }
        }
      ]
    }
  }
}
```

Each time you add a new extension to sunbeam, you gain access to new commands, and you can add new shortcuts to your config file.

> ℹ️ The source code of this extension is available here: <https://github.com/pomdtr/sunbeam-devdocs/blob/main/sunbeam-extension>.
