# Bitwarden (Bash / jq)

## Requirements

You will need to have the [Bitwarden CLI](https://bitwarden.com/help/article/cli/) and [jq](https://stedolan.github.io/jq/) installed.

The scripts require the `BW_SESSION` environment variable to be set to a valid session token.
Use the `bw login` command to generate a session token.

## Usage

```bash
sunbeam ./bitwarden.sh
```

## Code

```bash
{{#include ../code/bitwarden.sh}}
```
