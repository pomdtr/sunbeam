# Running the project

## Prerequisites

- Requires [Go](https://golang.org/dl/) 1.19 or later

## Setup Guide

### Dependencies

- Go >= 1.19

### Running the TUI

```bash
SUNBEAM_LOG_FILE=$PWD/debug.log go run main.go`
```

The logs will be redirected to the `debug.log` file, use `tail -f debug.log` to follow them.

## Installing the `sunbeam` command

```bash
go install
```
