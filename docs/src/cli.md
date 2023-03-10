# Command Line Interface

## Extending the default list

When run without arguments, sunbeam will read the default list from `~/.config/sunbeam/sunbeam.json`. You can extend this list by adding items to this file.

If your directory contains a `sunbeam.json` file, it will be used instead of the default list.

## Using sunbeam as a filter

There is two ways to wire sunbeam to your programs:

- Wrap your command in sunbeam: `sunbeam run <my-command>`
- Pipe data to sunbeam: `<my-command> | sunbeam read -`

Both methods have advantages and drawbacks:

- Using the wrap method, sunbeam controls the whole process. However, it requires your user to invoke sunbeam instead of your command, so you will loose completions.
- The pipe method is more flexible and easier to integrate with existing programs, but you loose the ability to reload your script since sunbeam is not responsible for generating the data.

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
