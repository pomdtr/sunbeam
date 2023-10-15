# Installation

## Requirements

A lot of extensions rely on having bash installed on your system.
MacOS and Linux devices comes with bash pre-installed, but you will need to install it on windows.

You can install bash on windows using [WSL](https://docs.microsoft.com/en-us/windows/wsl/install-win10) or [git bash](https://gitforwindows.org/).

## Installation

You can use a package manager

```bash
# using brew
brew install pomdtr/tap/sunbeam

# usign eget (https://github.com/zyedidia/eget)
eget install pomdtr/sunbeam --pre-release

# from source
go install github.com/pomdtr/sunbeam@latest
```

or use binaries / packages from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
