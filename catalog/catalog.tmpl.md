# Sunbeam Extensions

## Installing an Extension

1. Copy the URL in your clipboard
2. `sunbeam extension install <raw-url>`

## Catalog

{{catalog}}

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
