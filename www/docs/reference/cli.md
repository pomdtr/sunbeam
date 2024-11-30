---
outline: 2
---

# Cli

## sunbeam

Command Line Launcher

### Synopsis

Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.

```
sunbeam [flags]
```

### Options

```
  -h, --help   help for sunbeam
```

## sunbeam completion

Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for sunbeam for the specified shell.
See each sub-command's help for details on how to use the generated script.


### Options

```
  -h, --help   help for completion
```

## sunbeam completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(sunbeam completion bash)

To load completions for every new session, execute once:

#### Linux:

	sunbeam completion bash > /etc/bash_completion.d/sunbeam

#### macOS:

	sunbeam completion bash > $(brew --prefix)/etc/bash_completion.d/sunbeam

You will need to start a new shell for this setup to take effect.


```
sunbeam completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

## sunbeam completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	sunbeam completion fish | source

To load completions for every new session, execute once:

	sunbeam completion fish > ~/.config/fish/completions/sunbeam.fish

You will need to start a new shell for this setup to take effect.


```
sunbeam completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

## sunbeam completion help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type completion help [path to command] for full details.

```
sunbeam completion help [command] [flags]
```

### Options

```
  -h, --help   help for help
```

## sunbeam completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	sunbeam completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
sunbeam completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

## sunbeam completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(sunbeam completion zsh)

To load completions for every new session, execute once:

#### Linux:

	sunbeam completion zsh > "${fpath[1]}/_sunbeam"

#### macOS:

	sunbeam completion zsh > $(brew --prefix)/share/zsh/site-functions/_sunbeam

You will need to start a new shell for this setup to take effect.


```
sunbeam completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

## sunbeam copy

Copy text from stdin or paste text to stdout

```
sunbeam copy [flags]
```

### Options

```
  -h, --help   help for copy
```

## sunbeam edit

Open a file in your editor

```
sunbeam edit [file] [flags]
```

### Options

```
  -c, --config             Edit the config file
  -e, --extension string   File extension to use for temporary file
  -h, --help               help for edit
```

## sunbeam extension

Manage sunbeam extensions

### Options

```
  -h, --help   help for extension
```

## sunbeam extension configure

Configure extension preferences

```
sunbeam extension configure <alias> [flags]
```

### Options

```
  -h, --help   help for configure
```

## sunbeam extension create

Create a new sunbeam extension

```
sunbeam extension create <name> [flags]
```

### Options

```
  -h, --help              help for create
  -l, --language string   language of extension
```

## sunbeam extension help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type extension help [path to command] for full details.

```
sunbeam extension help [command] [flags]
```

### Options

```
  -h, --help   help for help
```

## sunbeam extension install

Install sunbeam extensions

```
sunbeam extension install <origin> [flags]
```

### Options

```
      --alias string   alias for extension
  -h, --help           help for install
```

## sunbeam extension list

List sunbeam extensions

```
sunbeam extension list [flags]
```

### Options

```
  -h, --help   help for list
```

## sunbeam extension publish

Publish a script as a github gist

```
sunbeam extension publish <script> [flags]
```

### Options

```
  -h, --help     help for publish
  -o, --open     open gist in browser
  -p, --public   make gist public
```

## sunbeam extension remove

Remove sunbeam extensions

```
sunbeam extension remove <alias> [flags]
```

### Options

```
  -h, --help   help for remove
```

## sunbeam extension rename

Rename sunbeam extensions

```
sunbeam extension rename <alias> <new-alias> [flags]
```

### Options

```
  -h, --help   help for rename
```

## sunbeam extension upgrade

Upgrade sunbeam extensions

```
sunbeam extension upgrade [flags]
```

### Options

```
      --all    upgrade all extensions
  -h, --help   help for upgrade
```

## sunbeam help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type sunbeam help [path to command] for full details.

```
sunbeam help [command] [flags]
```

### Options

```
  -h, --help   help for help
```

## sunbeam open

Open a file or url in your default application

```
sunbeam open [target] [flags]
```

### Options

```
  -h, --help   help for open
```

## sunbeam paste

Paste text from clipboard to stdout

```
sunbeam paste [flags]
```

### Options

```
  -h, --help   help for paste
```

## sunbeam query

Transform or generate JSON using a jq query

```
sunbeam query [query] [file] [flags]
```

### Options

```
      --arg stringArray       add string variable in the form of name=value
      --argjson stringArray   add JSON variable in the form of name=value
  -c, --compact-output        output without pretty-printing
  -h, --help                  help for query
  -i, --in-place              read and write to the same file
  -n, --null-input            use null as input value
  -R, --raw-input             read input as raw strings
  -r, --raw-output            output raw strings, not JSON texts
  -s, --slurp                 read all inputs into an array
      --yaml-input            read input as YAML format
      --yaml-output           output as YAML
```

## sunbeam validate

Validate a Sunbeam schema

### Options

```
  -h, --help   help for validate
```

## sunbeam validate config

Validate a config

```
sunbeam validate config [flags]
```

### Options

```
  -h, --help   help for config
```

## sunbeam validate detail

Validate a detail

```
sunbeam validate detail [flags]
```

### Options

```
  -h, --help   help for detail
```

## sunbeam validate help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type validate help [path to command] for full details.

```
sunbeam validate help [command] [flags]
```

### Options

```
  -h, --help   help for help
```

## sunbeam validate list

Validate a list

```
sunbeam validate list [flags]
```

### Options

```
  -h, --help   help for list
```

## sunbeam validate manifest

Validate a manifest

```
sunbeam validate manifest [flags]
```

### Options

```
  -h, --help   help for manifest
```

## sunbeam version

Print the version number of sunbeam

```
sunbeam version [flags]
```

### Options

```
  -h, --help   help for version
```


