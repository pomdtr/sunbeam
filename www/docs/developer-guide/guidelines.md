# Guidelines

## Choosing a language

Sunbeam extensions are just scripts, so you can use any language you want (as long as it can read and write JSON).

Sunbeam is not aware of the language you are using, so you will have to make sure that your script is executable and that it has the right shebang.

Even though you can use any language, here are some recommendations:

### POSIX Shell

Sunbeam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies (other than sunbeam itself).

- `sunbeam open`: open an url or a file using the default application.
- `sunbeam copy/paste`: copy/paste text from/to the clipboard

```sh
#!/bin/sh

set -eu

if [ $# -eq 0 ]; then
    jq -n '{
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
    jq -n '{ text: "Hello, World!" }'
fi
```

A more complex shell extension can be found [here](./examples/devdocs).

### Deno

[Deno](https://deno.land) is a secure runtime for javascript and typescript. It is an [excellent choice](https://matklad.github.io/2023/02/12/a-love-letter-to-deno.html) for writing scripts that require external dependencies.

Deno allows you to use any npm package by just importing it from a url. This makes it easy to use any library without requiring the user to install it first. The only requirement is that the user already has deno installed.

To make it easier to write extensions in deno, sunbeam provides a [deno package](https://deno.land/x/sunbeam) that provides types for validating the manifest and payloads. Make sure to use it to get the best experience.

```ts
#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "My Extension",
  description: "This is my extension",
  commands: [
    {
      name: "hi",
      title: "Say Hi",
      mode: "detail",
      params: [
        {
          name: "name",
          title: "Name",
          type: "text",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest
// the as const is required to make sure that the payload is correctly typed when parsing it
// the satisfies keyword allows you to get autocomplete/validation for the manifest

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

// here we pass the manifest type to the Payload type, so that the payload is correctly typed
const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "hi") {
  const name = payload.params.name;
  const detail: sunbeam.Detail = {
    text: `Hi ${name}!`,
    actions: [
      {
        title: "Copy Name",
        type: "copy",
        text: name,
      },
    ],
  };
  console.log(JSON.stringify(detail));
} else {
  console.error(`Unknown command: ${payload.command}`);
  Deno.exit(1);
}
```

A more complex typescript extension can be found [here](./examples/hackernews.md).

### Python

If you don't want to use deno/typescript ([you should really give it a try](https://matklad.github.io/2023/02/12/a-love-letter-to-deno.html)), you can use python instead.

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

Prefer to not use any external dependencies, as the user will have to install them manually. If you really need to use a dependency, you will need to distribute your extension through pip, and instruct the user how to install it.

See the [file-browser extension](./examples/file-browser.md) for an example.

### Any other language

You can use any language you want, as long as it can write/read JSON to/from stdout/stdin.
Just make sure to use the right shebang, or to compile your script to an binary executable.
