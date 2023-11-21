# hyperterm-sunbeam

Sunbeam integration for hyperterm.

## Installation

Use the Hyper CLI, bundled with your Hyper app, to install hyperterm-sunbeam
by entering the following into Hyper:

```bash
hyper i hyperterm-sunbeam
```

## Options

| Key          | Description                                             | Default  |
| ------------ | ------------------------------------------------------- | -------- |
| `hotkey`     | Shortcut<sup>1</sup> to toggle Hyper window visibility. | `Ctrl+;` |

## Example Config

```js
module.exports = {
  config: {
    sunbeam: {
      hotkey: "Alt+Super+O"
    }
  },
  plugins: ["hyperterm-sunbeam"]
};
```

<sup>1</sup> For a list of valid shortcuts, see [Electron Accelerators](https://github.com/electron/electron/blob/master/docs/api/accelerator.md).
