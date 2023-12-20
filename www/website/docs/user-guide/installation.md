# Installation

You can use a package manager

```bash
brew install pomdtr/tap/sunbeam

# install script
curl -sSf https://install.sunbeam.sh | sh

# as a nix flake
nix shell github:pomdtr/sunbeam --command sunbeam

# arch linux (btw)
yay -S sunbeam-bin

# from source
go install github.com/pomdtr/sunbeam@main
```

or use binaries / packages from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
