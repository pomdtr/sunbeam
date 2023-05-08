# Installation

## Requirements

- bash (used to run most script extensions)

On windows, you can get both by installing [git for windows](https://gitforwindows.org/).

## Installation

You can use a package manager

```bash
# brew (macos and linux)
brew install pomdtr/tap/sunbeam

# scoop (windows)
scoop bucket add pomdtr https://github.com/pomdtr/scoop-bucket.git
scoop install pomdtr/sunbeam

# install script (all platforms)
curl --proto '=https' --tlsv1.2 -LsSf https://raw.githubusercontent.com/pomdtr/sunbeam/main/scripts/install-sunbeam.sh | sh

# from source
go install github.com/pomdtr/sunbeam@latest
```

or use binaries / packages from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
