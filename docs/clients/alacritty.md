# Alacritty

[Alacritty](https://github.com/alacritty/alacritty) is a cross-platform terminal emulator.

You can use it as a sunbeam client with this config.

```yml
{{#include ./alacritty.yml}}
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
