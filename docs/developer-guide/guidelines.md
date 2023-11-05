# Guidelines

## Extension requirements

Your extension must print a valid manifest to stdout when invoked without arguments. See the [manifest](../reference/schemas/manifest.md) section for more information.

Your manifest describes a set of commands that can be invoked by the user. When a command is invoked, your extension will be called with a [payload](../reference/schemas/manifest.md) describing the command and its parameters.

If the mode of the command (as defined in the manifest) is either `list` or `detail`, your extension must print a valid [list](../reference/schemas/list.md) or [detail](../reference/schemas/detail.md) to stdout.

## Choosing a language

Sunbeam extensions are just scripts, so you can use any language you want (as long as it can read and write JSON).

Sunbeam is not aware of the language you are using, so you will have to make sure that your script is executable and that it has the right shebang.

Even though you can use any language, sunbeam provides multiple helpers to make it easier to write extensions in POSIX shell and deno.

### Shell

Sunbam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies (other than sunbeam itself).

- `sunbeam query`: generate and transform json using the jq syntax.
- `sunbeam fetch`: fetch a remote script using a subset of curl options.
- `sunbeam open`: open an url or a file using the default application.
- `sunbeam copy/paste`: copy/paste text from/to the clipboard

```sh
#! /bin/sh

# fail on error
set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Hello World!",
        commands: [{
            name: "say-hello",
            type: "string",
            required: true
        }]
    }'
    exit 0
fi

COMMAND=(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "say-hello" ]; then
    sunbeam query -n { text: "Hello, World!" }
fi
```

A more complex shell extension can be found [here](./shell).

### Deno

[Deno](https://deno.land) is a secure runtime for javascript and typescript. It is an [excellent choice](https://matklad.github.io/2023/02/12/a-love-letter-to-deno.html) for writing scripts that require external dependencies.

Deno allows you to use any npm package by just importing it from a url. This makes it easy to use any library without requiring the user to install it first.

To make it easier to write extensions in deno, sunbeam provides a [npm package](https://www.npmjs.com/package/sunbeam-types) that provides types for validating the manifest and payloads.

```ts
#! /usr/bin/env -S deno run -A

// import the types
import type * as sunbeam from "npm:sunbeam-types@0.23.16"

// import any npm package
// @deno-types="npm:@types/lodash"
import _ from "npm:lodash"

if (Deno.length == 0) {
    // if invoked without arguments, print the manifest
    const manifest: sunbeam.Manifest = {
        title: "Hello World!",
        commands: [{
            name: "say-hello",
            type: "string",
            required: true
        }]
    }

    Deno.exit(0)
}

// parse the payload
const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload

if (payload.command == "say-hello") {
    const detail: sunbeam.Detail = {
        text: _.uppercase("Hello, World!")
    }

    console.log(JSON.stringify(detail))
}
```

A more complex typescript extension can be found [here](./typescript.md).
