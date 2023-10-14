Sunbeam is a framework for writing keyboard-driven TUIs, by composing list, form and detail views. It's designed to be easy to use, and easy to extend.

### Sunbeam runs on all platforms

Sunbeam is distributed as a single binary, so you can run it on any platform. The sunbeam extension system is also designed to be cross-platform.

![sunbeam running in alacritty](./static/alacritty.png)

### Sunbeam is language agnostic

Sunbeam provides multiple helpers for writing scripts in POSIX shell, but you can also use any other language.

The only requirement is that your language of choice can read and write JSON.

Example Extensions:

- [VS Code](https://github.com/pomdtr/sunbeam-vscode)
- [File Browser](https://github.com/pomdtr/sunbeam-files)
- [Bitwarden](https://github.com/pomdtr/sunbeam-bitwarden)
- [Github](https://github.com/pomdtr/sunbeam-github)
- [TLDR Pages](https://github.com/pomdtr/sunbeam-tldr)
- [Devdocs](https://github.com/pomdtr/sunbeam-devdocs)

### Sunbeam is easy to extend

Instead of reiventing the wheel, sunbeam relies on your familiarity with git and github to make it easy to create, update, publish and install extensions.

Creating a new extension is as easy as writing a script. Sunbeam supports installing extension from any github repository, and can also interact with REST APIs.

Since sunbeam uses `git` under the hood, allowing you to create private extensions by using private git repositories.

![sunbeam running in vscode](./static/vscode.png)

### Sunbeam supports custom clients

Sunbeam comes with a built-in TUI to interact with your scripts, but you can also use any other client.

Currently the only alternative client is [sunbeam-raycast](https://github.com/pomdtr/sunbeam-raycast).

![raycast integration](./static/raycast.png)

## Inspirations / Alternatives

Sunbeam wouldn't exist without taking inspirations from incredible tools. Make sure to checkout:

- [raycast](https://raycast.com): Sunbeam shamelessly copy most of raycast UX. Even the project name is a reference to raycast.
- [fzf](https://github.com/junegunn/fzf): Sunbeam tries to take inspiration from fzf, but it's not a drop-in replacement. Sunbeam is designed to be used as a launcher, not as a fuzzy finder.
- [slapdash](https://slapdash.com): The sunbeam event loop was inspired by slapdash. Sadly, it looks like the slapdash development has been discontinued.
- [gh](https://cli.github.com): The sunbeam extension system was inspired by gh, with some modifications to make it more flexible.
