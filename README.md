Sunbeam is a TUI framework for creating keyboard-driven applications from simple scripts.

You can think of it as a mix between an application launcher like [raycast](https://raycast.com) and a fuzzy-finder like [fzf](https://github.com/junegunn/fzf).

[![asciicast](https://asciinema.org/a/614506.svg)](https://asciinema.org/a/614506)

## Features

## Runs on all platforms

Sunbeam is distributed as a single binary, available for all major platforms. Sunbeam also comes with a lot of utilities to make it easy to create cross-platform scripts.

![sunbeam running in alacritty](./static/alacritty.png)

## Supports any language

Sunbeam provides multiple helpers for writing scripts in POSIX shell, but you can also use any other language.

The only requirement is that your language of choice can read and write JSON.

Example Extensions:

- [VS Code (typescript)](https://github.com/pomdtr/sunbeam-vscode)
- [File Browser (python)](https://github.com/pomdtr/sunbeam-files)
- [Bitwarden (sh)](https://github.com/pomdtr/sunbeam-bitwarden)
- [Github (sh)](https://github.com/pomdtr/sunbeam-github)
- [TLDR Pages (sh)](https://github.com/pomdtr/sunbeam-tldr)
- [Devdocs (sh)](https://github.com/pomdtr/sunbeam-devdocs)

### Easy to extend

Instead of reiventing the wheel, sunbeam relies on your familiarity with git and github to make it easy to create, update, publish and install extensions.

Creating a new extension is as easy as writing a script.\
Sharing an extension is as easy as pushing it to github.

If you prefer, you can also distribute your extensions as http endpoints, and use sunbeam as a client.

![sunbeam running in vscode](./static/vscode.png)

### Bring your own UI

Sunbeam comes with a built-in TUI to interact with your scripts, but you can also use any other client.

See the client section for more details.

![raycast integration](./static/raycast.png)

## Inspirations / Alternatives

Sunbeam wouldn't exist without taking inspirations from incredible tools. Make sure to checkout:

- [raycast](https://raycast.com): Sunbeam shamelessly copy most of raycast UX. Even the project name is a reference to raycast.
- [fzf](https://github.com/junegunn/fzf): Sunbeam tries to take inspiration from fzf, but it's not a drop-in replacement. Sunbeam is designed to be used as a launcher, not as a fuzzy finder.
- [slapdash](https://slapdash.com): The sunbeam event loop was inspired by slapdash. Sadly, it looks like the slapdash development has been discontinued.
- [gh](https://cli.github.com): The sunbeam extension system was inspired by gh, with some modifications to make it more flexible.
