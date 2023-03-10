# CLI

## sunbeam

The `sunbeam` command is the main entry point to the sunbeam cli.

If a `sunbeam.json` file is present in the current directory, it will be used as the default page.

Otherwise, the default page will be read from the `$XDG_CONFIG_DIR/sunbeam/sunbeam.json` file. The file will be created if it doesn't exist.

## sunbeam run

Run the given command, and parse the output as a sunbeam page.

Accept a `--check` flag to non-interactively check if the command if the command output is valid.

### Examples

```bash
sunbeam run ./tldr.sh
sunbeam run -- ./file-browser.py --show-hidden
```

## sunbeam read

Read the given file, and parse the output as a sunbeam page.

Accept a `--check` flag to non-interactively check if the file is a valid sunbeam page.

### Examples

```bash
sunbeam read page.json
./file-browser.py --show-hidden | sunbeam read -
```

## sunbeam query

A wrapper around [jq](https://stedolan.github.io/jq/), useful to tranform json data into sunbeam pages.

See the [jq tutorial](https://stedolan.github.io/jq/tutorial/) to learn how to use jq.

> **Warning** The `sunbeam query` command slightly differs from the `jq` command. `jq --arg key value` is equivalent to `sunbeam query --arg key=value`.
