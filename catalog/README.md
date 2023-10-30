# Sunbeam Extensions

## Installing an Extension

1. Copy the URL in your clipboard
2. `sunbeam extension install <raw-url>`

## Catalog

| Extension                                                                                                | Description                     |
| -------------------------------------------------------------------------------------------------------- | ------------------------------- |
| [Mac Apps](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/macapps.sh)          | Open your favorite apps         |
| [File Browser](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/files.py)        | Browse files and folders        |
| [Tailscale](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/tailscale.ts)       | Manage your tailscale devices   |
| [Manage Extensions](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/manage.sh)  | Manage Sunbeam Extensions       |
| [GitHub](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/github.sh)             | Search GitHub Repositories      |
| [Gist](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/gist.ts)                 | Manage your gists               |
| [Quick Links](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/quicklinks.sh)    | Open your favorite websites     |
| [Bitwarden Vault](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/bitwarden.sh) | List your Bitwarden passwords   |
| [VS Code](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/vscode.ts)            | Manage your VS Code projects    |
| [Meteo](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/meteo.sh)               | Show Meteo                      |
| [System](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/system.sh)             | Control your system             |
| [Browse TLDR Pages](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/tldr.sh)    | Browse TLDR Pages               |
| [Edit Config Files](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/config.sh)  | Edit your favorite config files |
| [Val Town](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/valtown.sh)          | Manage your Vals                |
| [RSS](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/rss.ts)                   | Manage your RSS feeds           |
| [Cli Apps](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/apps.sh)             | Open your favorite cli apps     |
| [DevDocs](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/devdocs.sh)           | Search DevDocs.io               |
| [Raindrop](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/raindrop.ts)         | Manage your raindrop bookmarks  |

## Example

To install the devdocs extension, use:

```sh
# install the extension
sunbeam extension install https://raw.githubusercontent.com/pomdtr/sunbeam-extensions/main/extensions/devdocs.sh

# run the extension
sunbeam devdocs
```

Alternatively if you want to give it another name, pass the `--alias` flags

```sh
sunbeam extension install --alias docs https://raw.githubusercontent.com/pomdtr/sunbeam-extensions/main/extensions/devdocs.sh
sunbeam docs
```

You can also run the extension without installing it using:

```sh
sunbeam run https://raw.githubusercontent.com/pomdtr/sunbeam-extensions/main/extensions/devdocs.sh
```
