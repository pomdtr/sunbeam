# Bitwarden (Bash)

## Requirements

- [Bitwarden CLI](https://bitwarden.com/help/article/cli/)

> **Note** The scripts require the `BW_SESSION` environment variable to be set to a valid session token.
> Use the `bw login` command to generate a session token.

## Install

```bash
sunbeam extension install bw https://gist.github.com/pomdtr/bf9781444318cd9a5845444a6ac4f467
```

## Usage

```bash
sunbeam bw
```

## Code

```bash
{{#include ./sunbeam-extension}}
```
