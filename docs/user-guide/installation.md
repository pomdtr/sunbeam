# Installation

## Requirements

Sunbeam requires a POSIX shell (`sh`) to run. On windows, you can install one through [git bash](https://git-scm.com/downloads).

## Installation

You can use a package manager

```bash
# using brew (macOS / Linux)
brew install pomdtr/tap/sunbeam

# using scoop (Windows)
scoop bucket add pomdtr https://github.com/pomdtr/scoop-bucket.git
scoop install pomdtr/sunbeam

# install script
curl --proto '=https' --tlsv1.2 -LsSf https://github.com/pomdtr/sunbeam/releases/latest/download/install.sh | sh

# from source
go install github.com/pomdtr/sunbeam@latest

# usign eget (https://github.com/zyedidia/eget)
eget install pomdtr/sunbeam --pre-release
```

or use binaries / packages from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
