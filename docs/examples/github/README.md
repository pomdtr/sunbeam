# GitHub

## Requirements

- [gh](https://cli.github.com/)

> **Note** You should be authenticated with GitHub using the `gh auth login` command before running the scripts.

## Install

```bash
sunbeam extension install github https://pomdtr.github.io/sunbeam/book/examples/github/sunbeam-extension
```

## Usage

```bash
sunbeam github # List all repositories
sunbeam github list-prs <repo> # List all pull requests for a repository
```

## Code

```bash
{{#include ./sunbeam-extension}}
```