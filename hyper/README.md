# hyper-sunbeam

Use Hyper as a [Sunbeam](https://sunbeam.sh) GUI.

![sunbeam running in hyper](https://raw.githubusercontent.com/pomdtr/sunbeam/main/hyper/media/screenshot.jpeg)

## Installation

Use the Hyper CLI, bundled with your Hyper app, to install hyper-sunbeam:

```bash
hyper i hyper-sunbeam
```

## Options

| Key          | Description                                             | Default  |
| ------------ | ------------------------------------------------------- | -------- |
| `hotkey`     | Shortcut<sup>1</sup> to toggle Hyper window visibility. | `Ctrl+;` |

## Example Config

```js
module.exports = {
  config: {
    hyperSunbeam: {
      hotkey: "Alt+Super+O"
    }
  },
  plugins: ["hyper-sunbeam"]
};
```

<sup>1</sup> For a list of valid shortcuts, see [Electron Accelerators](https://github.com/electron/electron/blob/master/docs/api/accelerator.md).

## Suggested Extensions

Hyper Sunbeam couples well with the following extensions:

- [hyperborder](https://github.com/webmatze/hyperborder) - cool gradiant borders
- [hyper-transparent](https://github.com/codealchemist/hyper-transparent) - nice blur effect on macOS
