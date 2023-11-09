# CLI

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

## sunbeam fetch

Simple http client inspired by curl

```
sunbeam fetch <url> [body] [flags]
```

### Options

```
  -d, --data string          HTTP body to send. Use @- to read from stdin, or @<file> to read from a file.
  -H, --header stringArray   HTTP headers to add to the request
  -h, --help                 help for fetch
  -X, --method string        HTTP method to use
  -u, --user string          HTTP basic auth to use
  -A, --user-agent string    HTTP user agent to use
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
  -h, --help                  help for query
  -n, --null-input            use null as input value
  -R, --raw-input             read input as raw strings
  -r, --raw-output            output raw strings, not JSON texts
  -s, --slurp                 read all inputs into an array
```

## sunbeam run

Run an extension from a script, directory, or URL

```
sunbeam run <origin> [args...] [flags]
```

### Options

```
  -h, --help   help for run
```

## sunbeam upgrade



```
sunbeam upgrade [flags]
```

### Options

```
  -a, --all    upgrade all extensions
  -h, --help   help for upgrade
```

## sunbeam validate

Validate a Sunbeam schema

### Options

```
  -h, --help   help for validate
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


