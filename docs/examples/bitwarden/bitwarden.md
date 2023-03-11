# Bitwarden (Bash)

## Requirements

- [Bitwarden CLI](https://bitwarden.com/help/article/cli/)

> **Note** The scripts require the `BW_SESSION` environment variable to be set to a valid session token.
> Use the `bw login` command to generate a session token.

## Usage

```bash
sunbeam run ./bitwarden.sh
```

## Code

```bash
{{#include ./bitwarden.sh}}
```
