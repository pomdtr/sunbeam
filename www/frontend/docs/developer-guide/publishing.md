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

## Multiple File Extensions

Distribution of multiple file extensions is a bit more complicated, as sunbeam is not aware of the language you are using (it only understands json). If you can, consider publishing your extension as a single file, as it will be easier for your users to install it.

However, if you need to publish a multiple file extension, there are a few options available to you:

- If your extension is written in a compiled language, you can compile it to a single binary and publish it as a single file extension (ex: using github releases). Make sure to instruct your user to install the correct binary for their platform/architecture.
- If not, use the native package manager of your language (e.g. pip for python, npm for nodejs, etc.) to distribute your extension.

### Python

If your extension is written in python, you can publish it to [PyPI](https://pypi.org/). Make sure that the extension provides an `entry_points` in its `setup.py` file (or the equivalent in `pyproject.toml`).

Instruct your users to run `pip install --user <package>` to install it (or even better, use [pipx](https://pypa.github.io/pipx/)).

Now, they should be able to install your extension using `sunbeam extension install ~/.local/bin/<package>`.

> Note: If you don't want to publish your extension to PyPI, you can also install it from a git repository.
> `pip install git+https://github.com/<owner>/<package>.git`
