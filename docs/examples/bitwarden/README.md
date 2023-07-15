# Bitwarden (Bash)

## Requirements

- [Bitwarden CLI](https://bitwarden.com/help/article/cli/)

> **Note** The scripts require the `BW_SESSION` environment variable to be set to a valid session token.
> Use the `bw login` command to generate a session token.

## Install

```bash
sunbeam command add bw https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/bitwarden/bitwarden.sh
```

## Usage

```bash
sunbeam bw
```

## Code

```bash
{{#include ./bitwarden.sh}}
```
