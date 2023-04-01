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

## sunbeam copy

Copy stdin to system clipboard

```
sunbeam copy [flags]
```

### Options

```
  -h, --help   help for copy
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
  -h, --help   help for create
```

## sunbeam extension exec

Execute an installed extension

```
sunbeam extension exec [flags]
```

### Options

```
  -h, --help   help for exec
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

Install a sunbeam extension from a repository

```
sunbeam extension install [flags]
```

### Options

```
  -h, --help   help for install
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

## sunbeam extension rename

Rename an installed extension

```
sunbeam extension rename <old-name> <new-name> [flags]
```

### Options

```
  -h, --help   help for rename
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

## sunbeam fetch

fetch a page and push it's output

```
sunbeam fetch <url> [flags]
```

### Options

```
  -d, --data string          HTTP data
  -H, --header stringArray   HTTP header
  -h, --help                 help for fetch
  -X, --method string        HTTP method (default "GET")
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

Open file or url in default application

```
sunbeam open <url> [flags]
```

### Options

```
  -h, --help   help for open
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

Read page from file or stdin, and push it's content

```
sunbeam read [page] [flags]
```

### Options

```
  -f, --format string   Format of the input file. Can be json or yaml. (default "json")
  -h, --help            help for read
```

## sunbeam run

Run a script and push it's output

```
sunbeam run <script> [flags]
```

### Options

```
  -h, --help   help for run
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

Validate a page against the schema

```
sunbeam validate [file] [flags]
```

### Options

```
  -h, --help   help for validate
```


