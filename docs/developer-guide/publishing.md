# Publishing Extensions

## Single File Extensions

### Github Repositories

Sunbeam extensions can be published as file in a github repository.
Use the raw url of the file as the extension origin.

### Github Gists

If you prefer, you can also publish your extension as a gist.
Each file in the gist can be accessed throught its raw url.

### Github Releases

You can publish an extension script as a release asset.

To link to a release asset from the latest release, use the following url:

```
https://github.com/<owner>/<repo>/releases/latest/download/<script>
```

## Extension With Dependencies

If your extension has dependencies, you should publish it through the native package manager of your language (e.g. pip for python, npm for nodejs, etc.).

### Example - Python Apps

Make sure your extension is a valid python package, and publish it to PyPI.

Then, you can install it with pip:

```
pip install --user <package>
```

> Warning: Installing with pip at the system level is not recommended, as it may interfere with other packages.
> Consider using [pipx](https://pypa.github.io/pipx/) instead.

Once installed, you can use the path to the package as the extension origin in your sunbeam configuration.

```
{
    "extensions": {
        "<package>": "~/.local/bin/<package>"
    }
}
```

> Note: If you don't want to publish your extension to PyPI, you can also install it from a git repository.
> `pip install git+https://github.com/<owner>/<package>.git`
