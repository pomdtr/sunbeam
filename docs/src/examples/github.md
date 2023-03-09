# GitHub

## Requirements

You will need to have [gh](https://cli.github.com/) and [jq](https://stedolan.github.io/jq/) installed.

You should be authenticated with GitHub using the `gh auth login` command before running the scripts.

## Usage

```bash
sunbeam ./github.sh # List all repositories
sunbeam ./github.sh list-prs <repo> # List all pull requests for a repository
```

## Code

```bash
{{#include ../code/github.sh}}
```
