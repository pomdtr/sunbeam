# Bitwarden (Bash)

## Requirements

- [Bitwarden CLI](https://bitwarden.com/help/article/cli/)

> **Note** The scripts require the `BW_SESSION` environment variable to be set to a valid session token.
> Use the `bw login` command to generate a session token.

## Install

```bash
curl -L https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/bitwarden/bitwarden.bash > ~/.local/bin/sunbeam-bw
chmod +x ~/.local/bin/sunbeam-bw
```

## Usage

```bash
sunbeam bw list-passwords
```

## Code

```bash
{{#include ./bitwarden.bash}}
```
