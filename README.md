# Running the project

## Prerequisites

- go >= 1.19
- [direnv](https://direnv.net/): loads environment variables from .envrc (you can also directly source `.envrc`)

## Running the project

```console
direnv allow # load the SUNBEAM_COMMAND_DIR environment variable (you can also use `source .envrc`)
go run main.go
```

## Installing the `sunbeam` command

```console
go install
```
