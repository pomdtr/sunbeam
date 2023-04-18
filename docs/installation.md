# Installation

## Installation

You can use a package manager

```bash
# macOs or Linux
brew install pomdtr/tap/sunbeam

# Scoop
scoop bucket add pomdtr https://github.com/pomdtr/scoop-bucket.git
scoop install pomdtr/sunbeam
```

or install from source

```bash
go install github.com/pomdtr/sunbeam@latest
```

or download the binary from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
