# FAQ

## Using sunbeam as a filter

There is two ways to wire sunbeam to your programs:

- Run your command from sunbeam: `sunbeam exec ./my-command.sh`
- Pipe a page to sunbeam: `<my-command> | sunbeam read`
- Pipe rows to sunbeam: `gh repo list | sunbeam list`
- Pipe text to sunbeam: `echo "hello world" | sunbeam detail`
- Pipe an action to sunbeam: `<my-action> | sunbeam trigger`

## Configuring sunbeam appearance

You can configure the appearance of sunbeam by setting the following environment variables:

- `SUNBEAM_HEIGHT`: The maximum height of the sunbeam window, in lines. Defaults to `0` (full-height).
- `SUNBEAM_WIDTH`: The maximum width of the sunbeam window, in characters. Defaults to `0` (full-width).
- `SUNBEAM_BORDER`: Wether to display a border around the window. Defaults to `false`.
- `SUNBEAM_FULLSCREEN`: Wether to display the window in fullscreen. Defaults to `false`.

You can also set these options using the `--height` and `--padding` flags.

```bash
sunbeam ./github.sh
```

## Set Sunbeam Root View

By default, running `sunbeam` will display the usage string. You can set a root view to display instead by setting the `SUNBEAM_ROOT_CMD` environment variable.

```bash
SUNBEAM_ROOT_CMD="sunbeam extension manage" sunbeam
```

## Validating the output of a script

To validate the output of a script, you can use the validate command:

```bash
# Validate a static page
sunbeam validate sunbeam.json

# validate a dynamic page
./github.sh | sunbeam validate
```

The validate command will exit with a non-zero exit code if the output is invalid.

## I want to use sunbeam as a launcher

On macOS, you can integrate sunbeam with raycast using the [sunbeam raycast extension](https://github.com/pomdtr/sunbeam-raycast).

On Windows/Linux, there is no official integration yet, but you can [configure alacritty to launch sunbeam on startup](https://github.com/pomdtr/sunbeam/tree/main/assets/alacritty.yml), and use it as a launcher.

## Building a custom sunbeam frontend

Sunbeam use stdout to display pages, and stdin to receive actions.
If stdout is not interactive, sunbeam will dump the raw json to stdout instead.

You can use this to build your own frontend for sunbeam.

- Call `sunbeam trigger` with a root action, and read the output
- Display the page to the user
- If the user trigger an action, spawn a new sunbeam process with the JSON action as stdin
  - if the actions has inputs, you need to prompt the user for the input values, and pass them to sunbeam with the --input flag
