
# Sunbeam

![Demo](./demo/demo.gif)

> **Warning**: This is a work in progress. The API is not stable and may change at any time.

## Installation

```bash
go install github.com/pomdtr/sunbeam@latest
```

## Development

### Dependencies

- Go >= 1.19

### Running the TUI

```bash
SUNBEAM_LOG_FILE=$PWD/debug.log go run main.go`
```

The logs will be redirected to the `debug.log` file, use `tail -f debug.log` to follow them.
