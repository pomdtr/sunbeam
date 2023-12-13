# Publishing Extensions

## Single File Extensions

Single file extensions can be hosted anywhere, as long as they are accessible through a url.

You install them by adding the following snippet to your sunbeam config file:

```json
{
    "extensions": {
        "<name>": {
            "origin": "<url>"
        }
    }
}
```

Here are some examples of where you can host your extension. The list is not exhaustive.

### Github Repositories

Sunbeam extensions can be published as file in a github repository.
Use the raw url of the file as the extension origin.

### Github Gists

If you prefer, you can also publish your extension as a gist.
Each file in the gist can be accessed throught its raw url.

### Github Releases

You can publish an extension script as a release asset.

To link to a release asset from the latest release, use the following url:

```txt
https://github.com/<owner>/<repo>/releases/latest/download/<script>
```

## Extension With Dependencies

If your extension has dependencies, you should publish it through the native package manager of your language (e.g. pip for python, npm for nodejs, etc.), and instruct your users to run the appropriate sunbeam command to install it.

### Example for Python

If your extension is written in python, you can publish it to [PyPI](https://pypi.org/). Make sure that the extension provides an `entry_points` in its `setup.py` file.

Instruct your users to run `pip install --user <package>` to install it (or even better, use [pipx](https://pypa.github.io/pipx/)).

Now, they should be able to install your extension using `sunbeam extension install ~/.local/bin/<package>`.

> Note: If you don't want to publish your extension to PyPI, you can also install it from a git repository.
> `pip install git+https://github.com/<owner>/<package>.git`
