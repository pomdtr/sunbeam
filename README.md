
# Sunbeam

![Demo](./demo/demo.gif)

> **Warning**: This is a work in progress. The API is not stable and may change at any time.

## Roadmap

- Finalize API V1, Write Documentation
- An extension system inspired by [Github Cli](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions)
- Support for running the sunbeam host on a remote machine, and access it from the browser using xterm.js
- Add a GUI using wails.io

## Installation

```bash
go install github.com/SunbeamLauncher/sunbeam@latest
```

In order to install extension, either use the `sunbeam extension browse` command, or download one of the directory in the example folder and run `sunbeam extension install .`

## Development

## Libraries

- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema): For validating the config file and scripts output

### Dependencies

- Go >= 1.19

### Running the TUI

```bash
SUNBEAM_LOG_FILE=$PWD/debug.log go run main.go`
```

The logs will be redirected to the `debug.log` file, use `tail -f debug.log` to follow them.
