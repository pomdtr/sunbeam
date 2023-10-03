# Developer Guide

## Sunbeam Extensions

A sunbeam extension is a directory containing a `sunbeam-extension` file.

The file must respect the following contract:

- It must be executable (use `chmod +x <file>` to make it executable)
- It must return a sunbeam manifest when called with no arguments
- It must return a json payload on stdout when called with one argument (either a page or a command, depending on the manifest).

### Examples

Multiple examples are avaibles on github. Browse the [extension registry](https://github.com/topics/sunbeam-extension) to find them.

### Guides

TODO

## Publishing an extension

Create a git repository containing your extension, and push it to github.

Users can now install it using the `sunbeam extension install <repo-url>` command. Sunbeam will clone the repository on the user disk, and set an alias for the extension. The alias is the name of the repository, stripped of the `sunbeam-` prefix if it exists.

> ℹ️ sunbeam use `git` to clone the repository. If your repository is private, make sure that your user git credentials are set up correctly.

If you push new commits to the main branch, users can upgrade to the latest version using the `sunbeam extension upgrade <extension-alias>` command.

> ℹ️ Soon sunbeam will support more advanced publushing workflows, including pre-compiled extensions and tagged releases.

Then, add the `sunbeam-extension` topic to your repository. It will allow users to find your extension using the `sunbeam extension browse` command.
