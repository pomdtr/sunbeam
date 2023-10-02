# GitHub

## Requirements

- [gh](https://cli.github.com/)

> **Note** You should be authenticated with GitHub using the `gh auth login` command before running the scripts.

## Install

```bash
curl -L https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/github/github.bash > ~/.local/bin/sunbeam-gh
chmod +x ~/.local/bin/sunbeam-gh
```

## Usage

```bash
sunbeam gh list-repos # List all repositories
sunbeam github list-prs --repo excaldraw/excalidraw # List all pull requests for a repository
```

## Code

```bash
{{#include ./github.bash}}
```
