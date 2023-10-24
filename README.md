Sunbeam is a general purpose command-line launcher.

Define UIs composed of a succession of views from simple scripts written in any language.

<p align="center" style="text-align: center">
  <a href="https://asciinema.org/a/614506">
        <img src="https://asciinema.org/a/614506.svg">
  </a>
</p>

You can think of it as a mix between an application launcher like [raycast](https://raycast.com) or [rofi](https://github.com/davatorium/rofi) and a fuzzy-finder like [fzf](https://github.com/junegunn/fzf) or [telescope](https://github.com/nvim-telescope/telescope.nvim).

## Features

## Runs on all platforms

Sunbeam is distributed as a single binary, available for all major platforms. Sunbeam also comes with a lot of utilities to make it easy to create cross-platform scripts.

![sunbeam running in alacritty](./static/alacritty.png)

## Supports any language

Sunbeam provides multiple helpers for writing scripts in POSIX shell, but you can also use any other language.

The only requirement is that your language of choice can read and write JSON.

Example Extensions:

- [VS Code (typescript)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/vscode.ts)
- [File Browser (python)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/files.py)
- [Bitwarden (sh)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/bitwarden.sh)
- [Github (sh)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/github.sh)
- [TLDR Pages (sh)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/tldr.sh)
- [Devdocs (sh)](https://github.com/pomdtr/sunbeam-extensions/tree/main/extensions/devdocs.sh)

### Easy to extend

Creating a new extension is as easy as writing a script.
You can share your scripts with others by just hosting them on a public url.

![sunbeam running in vscode](./static/vscode.png)

### Bring your own UI

Sunbeam comes with a built-in TUI to interact with your scripts, but you can also use any other client.

See the client section for more details.

![raycast integration](./static/raycast.png)

## Inspirations / Alternatives

Sunbeam wouldn't exist without taking inspirations from incredible tools. Make sure to checkout:

- [raycast](https://raycast.com): Sunbeam shamelessly copy most of raycast UX, and nomenclature in it's api. Even the project name is a reference to raycast.
- [fzf](https://github.com/junegunn/fzf): Sunbeam tries to take inspiration from fzf, but it's not a drop-in replacement. Sunbeam is designed to be used as a launcher, not as a fuzzy finder.
- [slapdash](https://slapdash.com): Slapdash feature-set is quite close to sunbeam. Sadly, slapdash development seems to be stalled, and it's not open source.
- [gh](https://cli.github.com): Sunbeam extension system is taking inspiration from gh one.
