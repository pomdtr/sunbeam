# Running the project

## Prerequisites

- Requires [Go](https://golang.org/dl/) 1.19 or later
- Optional
  - [direnv](https://direnv.net/): loads environment variables from .envrc

## Setup Guide

### Github Codespaces

- Open the repository in github codespaces
- Run the cli: `go run main.go`

### Dev Container

- Install the Dev Container Extension
- Run `Dev Containers: Reopen in Container`
- Run the cli: `go run main.go`

### Local Install

- Install Dependencies:
  - go >= 1.19
  - direnv
- Allow `.envrc` config: `direnv allow`
- Run the cli: `go run main.go`

The logs are redirected to the `debug.log` file, use `tail -f debug.log` to follow them.

## Installing the `sunbeam` command

```console
go install
```
