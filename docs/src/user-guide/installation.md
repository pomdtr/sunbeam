# Installation

```sh
# go install (macOS, Linux and Windows)
go install github.com/pomdtr/sunbeam@latest
```

Sunbeam is a single binary, you can also download it from the [releases page](https://github.com/pomdtr/sunbeam/releases/latest).

## Configuring shell completions

Shell completions are available for bash, zsh and fish. To enable them, checkout the `sunbeam completion` command.

```sh
# bash
source <(sunbeam completion bash)

# zsh
source <(sunbeam completion zsh); compdef _sunbeam sunbeam

# fish
sunbeam completion fish | source
```
