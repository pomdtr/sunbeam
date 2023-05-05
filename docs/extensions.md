# Extending Sunbeam

The sunbeam extension system is heavily inspired by [gh](https://cli.github.com). Most of the [documentation](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) from gh can be applied to sunbeam.

## Using Extensions

Use the `sunbeam extension browse` commands to browse the sunbeam extensions available on github.
To install an extension, you can use the `sunbeam extension install <alias> <url>` command.

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

Any directory containing a `sunbeam-extension` executable is a valid sunbeam extension.
Use the `sunbeam extension create` command to bootstrap a new extension.

To test your extension, use the `sunbeam run ./sunbeam-extension` command, or the shorthand `sunbeam run .`.
You can install the current directory as an extension using the `sunbeam extension install <alias> .` command.

> **Warning**: Installing local extension is not yet supported on windows.

### Sunbeam Types

Types packages are available for the following languages:

- [go](https://pkg.go.dev/github.com/pomdtr/sunbeam/types)
- [typescript](https://www.npmjs.com/package/sunbeam-types)

## Distributing Extensions

You have three alternatives distribute sunbeam extensions:

### Github Gists

The easiest way to write a sunbeam extension is to create a [github gist](https://gist.github.com/) containing a `sunbeam-extension` executable.
You will then be able to run the extension using the `sunbeam run <gist-url>` command, or install it using the `sunbeam extension install <alias> <gist-url>` command.
All the examples from the docs are available as gists.

### Git Repositories

Alternatively, you can create a github repository containing the `sunbeam-extension` executable in it's root directory, and push it to github.
See the [sunbeam-deno-deploy](https://github.com/pomdtr/sunbeam-deno-deploy) extension for an example.

### Github Releases

Sunbeam also supports binary extensions.
In this case, sunbeam will download the binary from a github release instead of cloning the extension repository.

Your extension needs to respect the following naming convention: `sunbeam-extension-<os>-<arch>`.
The [sunbeam-extension-precompile](https://github.com/pomdtr/sunbeam-extension-precompile) github action can be used to automatically compile your extension for all systems and publish it as a github release.

See the [sunbeam-vscode](https://github.com/pomdtr/sunbeam-vscode) extension for an example.

## Publishing Extensions

> **Warning**: This feature is only available for extensions published as github repositories.

Set the `sunbeam-extension` topic on your repository to have it listed in the `sunbeam extension browse` command.
