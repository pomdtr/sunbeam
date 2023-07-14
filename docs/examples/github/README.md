# GitHub

## Requirements

- [gh](https://cli.github.com/)

> **Note** You should be authenticated with GitHub using the `gh auth login` command before running the scripts.

## Demo

![demo](./demo.gif)

## Install

```bash
sunbeam extension install github https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/github/github.sh
```

## Usage

```bash
sunbeam github # List all repositories
sunbeam github list-prs <repo> # List all pull requests for a repository
```

## Code

```bash
{{#include ./github.sh}}
```
