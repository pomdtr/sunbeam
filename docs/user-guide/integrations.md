# Integrations

Sunbeam is a built for the terminal first, but it can be used in other contexts. This section lists the available clients (outside of the TUI bundled with sunbeam).

## Raycast

[Raycast](https://raycast.com) is a macOS app that lets you control your tools with a few keystrokes.

You can use you sunbeam extension from raycast using the [sunbeam-raycast extension](https://github.com/pomdtr/sunbeam-raycast).

## Alacritty

[Alacritty](https://github.com/alacritty/alacritty) is a cross-platform terminal emulator.

You can use it as a sunbeam client with this config.

```yml
shell:
  program: /opt/homebrew/bin/fish
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

If you don't plan to use Alacritty as your primary terminal,
you can just save it as `~/.config/alacritty/alacritty.yml`.

```sh
# download the sunbeam config
curl https://pomdtr.github.io/sunbeam/book/clients/alacritty.yml > ~/.config/alacritty/alacritty.yml
# launch alacritty
alacritty
```

Otherwise, use the `config-file` flag when launching alacritty: `alacritty -c ~/.config/alacritty/sunbeam.yml`.

If you want to assign an hotkey to the alacritty window on MacOS, I highly recommend these tools:

- [raycast](https://www.raycast.com/)
- [skhd](https://github.com/koekeishiya/skhd)
- [hotkey](https://apps.apple.com/us/app/hotkey-app/id975890633?mt=12)

## Visual Studio Code

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

Trigger the keybinding and you should see the sunbeam menu appear in the terminal.
