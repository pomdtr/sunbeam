# Installation

## CLI

### homebrew tap

```console
brew install sunbeam
```

### eget

```console
eget sunbeamlauncher/sunbeam
```

### go install

```console
go install github.com/pomdtr/sunbeam@latest
```

### manually

Sunbeam is a single binary, you can download it from the [releases page](https://github.com/sunbeamlauncher/sunbeam/releases/latest).

## Configuring shell completions

Shell completions are available for bash, zsh and fish.

See the [completions page](./cmd/sunbeam_completion.md) for more information.

## GUI

Packages for the sunbeam GUI are available for Windows, Linux and MacOS on the [releases page](https://github.com/sunbeamlauncher//sunbeam-gui/releases/latest).

!!! info

    The GUI is not a wrapper around the CLI, it is a separate application that uses the sunbeam CLI to run commands.
    As such, you need to install the sunbeam CLI before installing the GUI.

!!! warning

    The GUI is still in early development, and is not required to use sunbeam.

### Post-installation steps

#### MacOS

Sunbeam is not notarized yet, so you will need to disable the quarantine before running the app or you will get an error. You can do this by running the following command in the terminal:

```console
xattr -d com.apple.quarantine /Applications/Sunbeam.app
```

#### Linux

Sunbeam is distributed as an AppImage. In order to run it, you will first need to make it executable.

```console
chmod +x Sunbeam-x.x.x.AppImage # Add execution permissions
./Sunbeam-x.x.x.AppImage # Run the application
```

Use the appimagelauncher tool to integrates the app in your desktop environment of choice.
