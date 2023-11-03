# Handling Dependencies

Sunbeam extensions are just scripts, so you can use any language you want to write them.

Sunebam provides multiple helpers to make it easier to share sunbeam extensions, without requiring the user to install additional dependencies.

- `sunbeam query`: generate and transform json using the jq syntax.
- `sunbeam fetch`: fetch a remote script using a subset of curl options.
- `sunbeam open`: open an url or a file using the default application.
- `sunbeam copy/paste`: copy/paste text from/to the clipboard

You can already build a lot of things using just these helpers, but sometimes you need to use a library to do something more complex.
In that case, consider dropping POSIX shell and using a more powerful language.

## Deno

[Deno](https://deno.land) is a secure runtime for javascript and typescript. It is an [excellent choice](https://matklad.github.io/2023/02/12/a-love-letter-to-deno.html) for writing scripts that require external dependencies.

Deno allows you to use any npm package by just importing it from a url. This makes it easy to use any library without requiring the user to install it first.

```ts
#!/usr/bin/env -S deno run -A

import { lodash } from "npm:lodash

// ...
```

To make it easier to write extensions in deno, sunbeam provides a [npm package](https://www.npmjs.com/package/sunbeam-types) that provides types for validating the manifest and payloads.

## Nix

[Nix](https://nixos.org) is a purely functional package manager. Among a ton of other things, it allows you to write portable scripts by leveraging shebangs.

```sh
#! /usr/bin/env nix-shell
#! nix-shell -i bash -p imagemagick cowsay

# scale image by 50%
convert "$1" -scale 50% "$1.s50.jpg" &&
cowsay "done $1.q50.jpg"
```

See the [nix manual](https://nixos.wiki/wiki/Nix-shell_shebang) for more information.

This [blog post](https://yukiisbo.red/notes/spice-up-with-nix-scripts/) also provides a good introduction to writing nix-powered scripts.
