# Integrations

Sunbeam is a terminal-based application, so it can be integrated with any terminal, multiplexer or editor.

Here is a non-exhaustive list of integrations. If you have an integration you would like to share, feel free to open a PR.

## Terminals

### Hyper

![Sunbeam running in Hyper](./../../assets/hyper.jpeg)

[Hyper](https://hyper.is/) is an extensible terminal emulator, built on web technologies.
It is not the lighter terminal emulator (far from it), but it is easy to extend and customize.

The [hyper-sunbeam](https://npm.js/hyper-sunbeam) plugin allows you to transform hyper into a sunbeam launcher.

To install it, just create an `.hyper.js` file in your home directory:

```js
"use strict";

// See https://hyper.is#cfg for all currently supported options.
module.exports = {
    config: {
        fontSize: 13,
        padding: '10px',
        shellArgs: ['-lic', 'sunbeam'],
        // these envs are used by sunbeam actions, make sure to customize them if needed
        env: {
            "EDITOR": "vim"
        },
        windowSize: [750, 440],
        modifierKeys: {
            altIsMeta: true // needed for hotkeys to work
        },
        hyperSunbeam: {
            hotkey: 'Shift+Super+K', // setup your hotkey here
        }
    },
    // a list of plugins to fetch and install from npm
    plugins: [
        "hyper-sunbeam", // tranform hyper into a floating panel
        "hyper-monokai-pro", // suggested theme, but you can use any theme you want
    ]
};
```

Feel free to customize the config to your liking.

> ⚠️
> If you're desktop environment is using wayland, global hotkeys will not work. Instead, setup an hotkey in your desktop settings to launch hyper. Launching hyper while it is already running will toggle it instead of opening a new window.

### Alacritty

![Sunbeam running in alacritty](./../../assets/alacritty.jpeg)

[Alacritty](https://github.com/alacritty/alacritty) is a cross-platform terminal emulator.

Alacritty is not easily extensible, so you will have to handle the application launcher features yourself (hotkey, centering, blur, ect.).

It is a good choice if you are already using tiling window manager, as they usually allow you to setup an hotkey to launch a program, and to center it on the screen.

Use this config as a starting point:

```yml
env:
  EDITOR: vim
shell:
  program: /bin/bash # set this to your shell
  args: ["-lic", "sunbeam"]

window:
  opacity: 0.9
  decorations: none
  dimensions:
    columns: 90
    lines: 23

  option_as_alt: Both
  padding:
    x: 10
    y: 10

font:
  size: 13.0
```

On Gnome, you can configure new windows to be centered on the screen by running `gsettings set org.gnome.mutter center-new-windows true`. Then just assign a hotkey to launch alacritty.

If you don't plan to use Alacritty as your primary terminal, you can just save it as `~/.config/alacritty/alacritty.yml`.
Otherwise, use the `config-file` flag when launching alacritty: `alacritty --config-file ~/.config/alacritty/sunbeam.yml`.

### Tweety

![Sunbeam running in tweety](./../../assets/tweety.jpeg)
Tweety is an integrated terminal for your web browser.

It can be embedded anywhere you can show a web page, making it a great GUI for sunbeam.

Example configuration:

```json
{
  "env": {
    "EDITOR": "vim",
  },
  "defaultProfile": "sunbeam",
  "profiles": {
    "sunbeam": {
      "command": "/bin/zsh",
      "args": [
        "-c",
        "sunbeam"
      ]
    }
  }
}
```

## Multiplexers

Sunbeam can easily be integrated with terminal multiplexers like tmux or zellij.

### tmux

```sh
tmux popup -E sunbeam # open sunbeam in a popup
tmux display-popup -E sunbeam devdocs list-docsets # list devdocs docsets in a popup
```

To bind it to a key, add this line to your tmux config:

```sh
bind-key -n C-Space display-popup -E sunbeam
```

### zellij

```sh
zellij run --floating --close-on-exit -- htop
```

Binding this command to a key is not supported yet, as zellij [does not support floating panes in its config file yet](https://github.com/zellij-org/zellij/discussions/2518).

## Editors

### Visual Studio Code

![Sunbeam running in VS Code](./../../assets/vscode.png)

Run the `Tasks: Open User Tasks` item from the command palette, then paste the following text:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "type": "shell",
            "label": "sunbeam",
            "command": "sunbeam",
            "presentation": {
                "echo": false,
                "focus": true,
                "close": true
            }
        }
    ]
}
```

Then run the `Preferences: Open Default Keyboard Shortcuts (JSON)` command and add the following keybinding to the list:

```json
{
    "key": "ctrl+alt+p",
    "command": "workbench.action.tasks.runTask",
    "args": "sunbeam"
}
```

Trigger the keybinding and you should see the sunbeam menu appear in the terminal panel.

### Kakoune

You can use the [popup.kak](https://github.com/enricozb/popup.kak) plugin to show sunbeam in a popup.

```sh
evaluate-commands %sh{kak-popup init}

define-command -override -params .. sunbeam %{ popup --title open -- sunbeam %arg{@} }

map global user <space> ':sunbeam<ret>' -docstring "Show Sunbeam"
```

### Vim / Neovim

Checkout the following plugins:

- [floaterm](https://github.com/voldikss/vim-floaterm)
- [FTerm.nvim](https://github.com/numToStr/FTerm.nvim)
- [toggleterm.nvim](https://github.com/akinsho/toggleterm.nvim)

or just use the `:terminal` command.

## Shells

### Fish

To bind sunbeam to a key, use the `bind` command:

```fish
# bind sunbeam to ctrl+space
bind -k nul 'sunbeam'
```

## GUI (TODO)

A sunbeam GUI is in the works, but it is not ready yet.

It will be close to the current hyper integration, but available as a standalone app.
