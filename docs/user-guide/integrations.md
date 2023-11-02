# Integrations

Sunbeam is a built for the terminal first, but it can be used in other contexts. This section lists the available clients (outside of the TUI bundled with sunbeam).

## Terminals

### Hyper

[Hyper](https://hyper.is/) is a cross-platform terminal emulator, built on web technologies.

You can use the [sunbeam plugin](https://www.npmjs.com/package/hyper-sunbeam) to make hyper behave as an application launcher.
The plugin has been tested on macOS, but it should work on other platforms as well.

Here is is my config:

```js
"use strict";
// See https://hyper.is#cfg for all currently supported options.
module.exports = {
    config: {
        // default font size in pixels for all tabs
        fontSize: 13,
        // custom padding (CSS format, i.e.: `top right bottom left`)
        padding: '10px 0px 5px 5px',
        // the shell to run when spawning a new session (i.e. /usr/local/bin/fish)
        // if left empty, your system's login shell will be used by default
        shell: '/opt/homebrew/bin/fish',
        // for setting shell arguments (i.e. for using interactive shellArgs: `['-i']`)
        // by default `['--login']` will be used
        shellArgs: ['--login', '-c', 'sunbeam'],
        // for environment variables
        env: {
            "EDITOR": "kak"
        },
        windowSize: [600, 350],
        // for advanced config flags please refer to https://hyper.is/#cfg
        modifierKeys: {
            altIsMeta: true
        },
        sunbeam: {
            hotkey: 'Alt+Super+Shift+Control+Space',
        },
        hypest: {
            vibrancy: true,
            hideControls: true,
            darkmode: true,
            borders: true
        }
    },
    // a list of plugins to fetch and install from npm
    plugins: [
        "hyper-sunbeam",
        "hyperminimal",
        "hyperborder",
        "hyper-hypest",
    ]
};
//# sourceMappingURL=config-default.js.map
```

### Alacritty

[Alacritty](https://github.com/alacritty/alacritty) is a cross-platform terminal emulator.

Alacritty is not easily extensible, so you will have to handle the application launcher features yourself (hotkey, centering, blur, ect.).

It is a good choice if you use a tiling window manager, as they usually have built-in support for advanced window management.

Use this config as a starting point:

```yml
shell:
  program: /bin/bash
  args: ["-lic", "sunbeam"]

window:
  opacity: 0.9
  decorations: buttonless
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

If you don't plan to use Alacritty as your primary terminal, you can just save it as `~/.config/alacritty/alacritty.yml`.
Otherwise, use the `config-file` flag when launching alacritty: `alacritty -c ~/.config/alacritty/sunbeam.yml`.

## Editors

### Visual Studio Code

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
