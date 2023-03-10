# GitHub

## Requirements

You will need to have [gh](https://cli.github.com/) and [jq](https://stedolan.github.io/jq/) installed.

You should be authenticated with GitHub using the `gh auth login` command before running the scripts.

## Usage

```bash
sunbeam run ./github.sh # List all repositories
sunbeam run ./github.sh list-prs <repo> # List all pull requests for a repository
```

## Code

```bash
{{#include ../code/github.sh}}
```
