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

## sunbeam extension

Extension commands

### Options

```
  -h, --help   help for extension
```

## sunbeam extension browse

Browse extensions

```
sunbeam extension browse [flags]
```

### Options

```
  -h, --help   help for browse
```

## sunbeam extension create

Create a new extension

```
sunbeam extension create [flags]
```

### Options

```
  -h, --help          help for create
  -n, --name string   Extension name
```

## sunbeam extension install

Install a sunbeam extension from a repository

```
sunbeam extension install [flags]
```

### Options

```
  -h, --help   help for install
  -o, --open   Open extension after installation
```

## sunbeam extension list

List installed extension commands

```
sunbeam extension list [flags]
```

### Options

```
  -h, --help   help for list
```

## sunbeam extension manage

Manage installed extensions

```
sunbeam extension manage [flags]
```

### Options

```
  -h, --help   help for manage
```

## sunbeam extension remove

Remove an installed extension

```
sunbeam extension remove [flags]
```

### Options

```
  -h, --help   help for remove
```

## sunbeam extension search

Search for repositories with the sunbeam-extension topic

```
sunbeam extension search [flags]
```

### Options

```
  -h, --help   help for search
```

## sunbeam extension upgrade

Upgrade an installed extension

```
sunbeam extension upgrade [flags]
```

### Options

```
  -h, --help   help for upgrade
```

## sunbeam extension view

View extension

```
sunbeam extension view <repo> [flags]
```

### Options

```
  -h, --help   help for view
```

## sunbeam query

Transform or generate JSON using a jq query

```
sunbeam query <query> [flags]
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

## sunbeam read

Read page from file, and push it

```
sunbeam read <page> [flags]
```

### Options

```
  -h, --help   help for read
```

## sunbeam run

Generate a page from a command or a script, and push it

```
sunbeam run [flags]
```

### Options

```
  -h, --help                help for run
      --on-success string   action to trigger when the command is successful (default "push")
```

## sunbeam trigger

Trigger an action

```
sunbeam trigger <action> [flags]
```

### Options

```
  -h, --help                 help for trigger
      --inputs stringArray   inputs to pass to the action
```

## sunbeam validate

Validate a page

```
sunbeam validate [file] [flags]
```

### Options

```
  -h, --help   help for validate
```


