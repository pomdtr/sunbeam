# Running the project

## Prerequisites

- Requires [Go](https://golang.org/dl/) 1.19 or later
- Optional
  - [direnv](https://direnv.net/): loads environment variables from .envrc
  - [jo](https://github.com/jpmens/jo): creates JSON objects from the command line (useful for testing scripts)

## Running the project

```console
direnv allow # load the SUNBEAM_COMMAND_DIR environment variable (you can also use `source .envrc`)
go run main.go
```

The logs are redirected to the `debug.log` file, use `tail -f debug.log` to follow them.

## Installing the `sunbeam` command

```console
go install
```
