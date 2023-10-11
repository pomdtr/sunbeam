# Developer Guide

Sunbeam extensions can be written in any language.
They are executed as standalone processes, and communicate with the sunbeam using json payloads.

However, some languages are better suited than others for writing sunbeam extensions:

- shell: Shell scripts are a good choice for writing simple sunbeam extensions. They are easy to write, and can be executed directly by sunbeam.
    - `sunbeam query` allows you to create/transfor json payloads (it's a go port of [jq](https://stedolan.github.io/jq/))
    - `sunbeam fetch` allows you to perform http requests (with an api similar to [curl](https://curl.se/))
- deno: Deno is a secure runtime for javascript/typescript. It's a good choice for writing complex sunbeam extensions, since importing external dependencies can be done using a url.

## Anatomy of an extension

A sunbeam extension is a directory containing a `sunbeam-extension` file.

The file must respect the following contract:

- It must be executable (use `chmod +x <file>` to make it executable)
- When called without arguments, it must return a json manifest describing the extension and its commands.
- When called with a command name as first argument, it must execute the command, and optionally return a json payload on stdout.
    - The output schema depends on the command mode defined in the manifest.

## Releasing an extension

Create a git repository containing your extension, and push it to github.

Users can now install it using the `sunbeam extension install <repo-url>` command. Sunbeam will clone the repository on the user disk, and set an alias for the extension. The alias is the name of the repository, stripped of the `sunbeam-` prefix if it exists.

> ℹ️ sunbeam use `git` to clone the repository. If your repository is private, make sure that your user git credentials are set up correctly.

If you push new commits to the main branch, users can upgrade to the latest version using the `sunbeam extension upgrade <extension-alias>` command.

If you prefer, you can use tags instead. Sunbeam will use the latest tag (according to semver) as the extension version.

## Publishing an extension

Add the `sunbeam-extension` topic to your github repository to make it discoverable.

You can use the `sunbeam extension browse` command from sunbeam to open the extension registry in your browser.
