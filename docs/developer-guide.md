# Developer Guide

Sunbeam extensions can be written in any language. They are executed as standalone processes, and communicate with the sunbeam core using json payloads. However, some languages are better suited than others for writing sunbeam extensions.

- Bash: sunbeam provides helpers to write bash extensions.
    - `sunbeam query` allows you to create/transfor json payloads (it's a go port of [jq](https://stedolan.github.io/jq/))
    - `sunbeam fetch` allows you to perform http requests (with an api similar to [curl](https://curl.se/))
- Deno: Deno is a secure runtime for javascript/typescript. It's a good choice for writing sunbeam extensions with external dependencies.
- Go: Go is a compiled language. It's a good choice for writing sunbeam extensions that need to be fast and portable.

## Example Extensions

Take a look at the extension available in the [extension registry](https://github.com/topics/sunbeam-extension) to get a better idea of what a sunbeam extension looks like.

## Sunbeam Extensions

A sunbeam extension is a directory containing a `sunbeam-extension` file.

The file must respect the following contract:

- It must be executable (use `chmod +x <file>` to make it executable)
- When called without arguments, it must return a json manifest describing the extension and its commands.
- When called with a command name as first argument, it must execute the command and return a json payload.
    - The command parameters are passed as json payload on stdin

### Manifest

### Input payload

### Output payload

### Examples

Multiple examples are avaibles on github. Browse the [extension registry](https://github.com/topics/sunbeam-extension) to find them.

## Publishing an extension

Create a git repository containing your extension, and push it to github.

Users can now install it using the `sunbeam extension install <repo-url>` command. Sunbeam will clone the repository on the user disk, and set an alias for the extension. The alias is the name of the repository, stripped of the `sunbeam-` prefix if it exists.

> ℹ️ sunbeam use `git` to clone the repository. If your repository is private, make sure that your user git credentials are set up correctly.

If you push new commits to the main branch, users can upgrade to the latest version using the `sunbeam extension upgrade <extension-alias>` command.

> ℹ️ Soon sunbeam will support more advanced publushing workflows, including pre-compiled extensions and tagged releases.

Then, add the `sunbeam-extension` topic to your repository. It will allow users to find your extension using the `sunbeam extension browse` command.
