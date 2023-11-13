# Guidelines

## Choosing a language

Sunbeam extensions are just scripts, so you can use any language you want (as long as it can read and write JSON).

Sunbeam is not aware of the language you are using, so you will have to make sure that your script is executable and that it has the right shebang.

Even though you can use any language, here are some recommendations:

### POSIX Shell

Sunbam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies (other than sunbeam itself).

- `sunbeam query`: generate and transform json using the jq syntax.
- `sunbeam fetch`: fetch a remote script using a subset of curl options.
- `sunbeam open`: open an url or a file using the default application.
- `sunbeam copy/paste`: copy/paste text from/to the clipboard

```sh
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    sunbeam query -n '{
        title: "Hello World!",
        commands: [{
            name: "say-hello",
            title: "Say Hello",
            mode: "detail"
        }]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
if [ "$COMMAND" = "say-hello" ]; then
    sunbeam query -n '{ text: "Hello, World!" }'
fi
```

A more complex shell extension can be found [here](./examples/devdocs).

### Python 3

If your shell script is getting too complex, consider rewriting it in python.

Python3 comes preinstalled in macOS and on most linux distributions, so it is a good choice if you want to write an extension that can be used without requiring the user to install additional dependencies.

Make sure to use the `#!/usr/bin/env python3` shebang, as it will make your script more portable.

```python
#!/usr/bin/env python3

import sys
import json

if len(sys.argv) == 1:
    manifest = {
        "title": "Hello World!",
        "commands": [{
            "name": "say-hello",
            "title": "Say Hello",
            "mode": "detail"
        }]
    }

    print(json.dumps(manifest))
    sys.exit(0)

payload = json.loads(sys.argv[1])
if payload["command"] == "say-hello":
    detail = {
        "text": "Hello, World!"
    }

    print(json.dumps(detail))
```

Prefer to not use any external dependencies, as the user will have to install them manually.

See the [file-browser extension](./examples/file-browser.md) for an example.

### Deno

[Deno](https://deno.land) is a secure runtime for javascript and typescript. It is an [excellent choice](https://matklad.github.io/2023/02/12/a-love-letter-to-deno.html) for writing scripts that require external dependencies.

Deno allows you to use any npm package by just importing it from a url. This makes it easy to use any library without requiring the user to install it first. The only requirement is that the user already has deno installed.

To make it easier to write extensions in deno, sunbeam provides a [npm package](https://www.npmjs.com/package/sunbeam-types) that provides types for validating the manifest and payloads.

```ts
#!/usr/bin/env -S deno run -A

// import the types
import type * as sunbeam from "npm:sunbeam-types@0.23.18"

// import any npm package
// @deno-types="npm:@types/lodash"
import _ from "npm:lodash"

if (Deno.args.length == 0) {
    // if invoked without arguments, print the manifest
    const manifest: sunbeam.Manifest = {
        title: "Hello World!",
        commands: [{
            name: "say-hello",
            title: "Say Hello",
            mode: "detail",
        }]
    }

    console.log(JSON.stringify(manifest))
    Deno.exit(0)
}

// parse the payload
const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload

if (payload.command == "say-hello") {
    const detail: sunbeam.Detail = {
        text: _.upperCase("Hello, World!")
    }

    console.log(JSON.stringify(detail))
}

```

A more complex typescript extension can be found [here](./examples/hackernews.md).
