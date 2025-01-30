# Installation

## Package Manager

### Homebrew

```bash
brew install pomdtr/tap/sunbeam
```

### Nix

```bash
nix shell github:pomdtr/sunbeam --command sunbeam
```

### Yay (Arch Linux, btw)

```bash
yay -S sunbeam-bin
```

## From Source

```bash
go install github.com/pomdtr/sunbeam@main
```

## GitHub Releases

Alternatively, use binaries / packages from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Completions

Sunbeam supports completions for bash, zsh, fish and powershell. If you installed sunbeam using a package manager, the completions should be installed automatically. Otherwise, you can install them manually.

Run `sunbeam completion <shell> --help` to list the available options.
