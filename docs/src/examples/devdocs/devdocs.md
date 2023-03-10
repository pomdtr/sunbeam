# Devdocs (Bash / jq)

## Requirements

You will need to have [curl](https://curl.haxx.se/) and [jq](https://stedolan.github.io/jq/) installed.

## Usage

```bash
sunbeam run ./devdocs.sh # List all docsets
sunbeam run ./devdocs.sh <docset-slug> # List all entries for a docset
```

## Code

```bash
{{#include ./devdocs.sh}}
```
