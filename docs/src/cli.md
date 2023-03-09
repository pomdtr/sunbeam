# Command Line Interface

## Configuring sunbeam appearance

You can configure the appearance of sunbeam by setting the following environment variables:

- `SUNBEAM_HEIGHT`: The maximum height of the sunbeam window, in lines. Defaults to `0` (fullscreen).
- `SUNBEAM_PADDING`: The padding around the sunbeam window. Defaults to `0`.

You can also set these options using the `--height` and `--padding` flags.

```bash
sunbeam --height 20 --padding 2 ./github.sh
```

## Validating the output of a script

To validate the output of a script, you can use the `--check` command:

```bash
sunbeam --check ./github.sh
```

The interactive UI will not be shown, but the output will be validated and an error will be thrown if the output is not valid.
