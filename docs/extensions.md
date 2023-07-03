# Extending Sunbeam

The sunbeam extension system is heavily inspired by [gh](https://cli.github.com). Most of the [documentation](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) from gh can be applied to sunbeam.

## Using Extensions

Use the `sunbeam extension browse` commands to browse the sunbeam extensions available on github.
To install an extension, you can use the `sunbeam extension install <name> <url>` command.

For example, to install the `sunbeam-vscode` extension as `vsc`:

```bash
# Install the extension
sunbeam extension install vsc https://github.com/pomdtr/sunbeam-vscode

# Run the extension
sunbeam vsc
```

The `subeam extension manage` command can be used to manage installed extensions.
Alternatively, use the `list`, `remove`, `upgrade` and `rename` commands directly.

## Writing Extensions

Any directory containing a `sunbeam-command` executable is a valid sunbeam extension.

To test your extension, use the `sunbeam run ./sunbeam-command` command.
You can install the current directory as an extension using the `sunbeam extension install <alias> .` command.

> **Warning**: Installing local extension is not yet supported on windows.

You can write extension using any language. If you want to distribute your extension, make sure that you provide instructions on how to install the required dependencies.

Here are some suggestions if you don't know what language to use:

- Bash is already installed on most systems. Sunbeam provides multiple commands to help you write bash extensions.
  - sunbeam argparse: parse command line arguments
  - sunbeam require: check if a command is available
  - sunbeam query: Transform or generate JSON
  - sunbeam fetch: Perform HTTP requests
- Go is a good choice for more complex extensions. The [sunbeam/types](https://pkg.go.dev/github.com/pomdtr/sunbeam/types) package provides types for all sunbeam commands. Go binaries can be distributed as github releases.
- If you are more confortable with javascript/typescript, take a look at [deno](https://deno.land/). Types are available both on [npm](https://npmjs.com/package/sunbeam-types) and [deno.land](https://deno.land/x/sunbeam/index.d.ts).

Use the `sunbeam extension create` command to bootstrap a new extension.

## Distributing Extensions

You have three alternatives distribute sunbeam extensions:

### Github Gists

The easiest way to write a sunbeam extension is to create a [github gist](https://gist.github.com/) containing a `sunbeam-command` executable.
You will then be able to run the extension using the `sunbeam run <gist-url>` command, or install it using the `sunbeam extension install <name> <gist-url>` command.

All the examples from the docs are available as gists.

### Git Repositories

Alternatively, you can create a github repository containing the `sunbeam-command` executable in it's root directory, and push it to github.
See the [sunbeam-deno-deploy](https://github.com/pomdtr/sunbeam-deno-deploy) extension for an example.

### Github Releases

Sunbeam also supports binary extensions.
In this case, sunbeam will download the binary from a github release instead of cloning the extension repository.

Your extension needs to respect the following naming convention: `sunbeam-command-<os>-<arch>`.
The [sunbeam-command-precompile](https://github.com/pomdtr/sunbeam-command-precompile) github action can be used to automatically compile your extension for all systems and publish it as a github release.

See the [sunbeam-vscode](https://github.com/pomdtr/sunbeam-vscode) extension for an example.

## Publishing Extensions

> **Warning**: This feature is only available for extensions published as github repositories.

Set the `sunbeam-command` topic on your repository to have it listed in the `sunbeam extension browse` command.

Be sure to test your extension before publishing it.
