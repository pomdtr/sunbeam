# Command Line Interface

## Detecting that a script is running in sunbeam

Sunbeam set the `SUNBEAM_RUNNER` environment variable to `true` when it's running a script. You can use it to adapt the output of your script depending on the context.

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
sunbeam run --check ./github.sh
sunbeam read --check sunbeam.json
```

The interactive UI will not be shown, but the output will be validated. If the output is invalid, the command will exit with a non-zero status code.
