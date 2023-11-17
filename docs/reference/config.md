# Config

## Static Config

The config will be searched in the following locations:

- `$SUNBEAM_CONFIG_DIR/sunbeamrc` if `$SUNBEAM_CONFIG_DIR` if set
- `$XDG_CONFIG_HOME/sunbeam/sunbeamrc` if `XDG_CONFIG_HOME` is set
- `$HOME/.config/sunbeam/config.json`

If no config is found, sunbeam will create one.

```json
{
    // additional items to show in the root list
    "oneliners": {
            "View System Resources": "htop"
    },
    "extensions": {
        "github": {
            "origin": "~/Developer/github.com/pomdtr/sunbeam/extensions/github.sh",
            // preferences for the extension, use it to pass config or secrets
            "preferences": {
                "token": "xxxx"
            },
            // additional root items to show
            "items": [
                {
                    "title": "List Sunbeam Issues",
                    "command": "list-issues",
                    "params": {
                        "repo": "pomdtr/sunbeam",
                    }
                }
            ]
        }
    }
}
```

## Dynamic Config

If you want to have more control over the root items, you can use a script to generate the config.

Create a script in in named `sunbeamrc` in the config directory. The script will be executed every time sunbeam is opened, and the output will be used as the config.

```bash
#!/bin/sh

cat <<EOF
{
    // additional items to show in the root list
    "oneliners": {
            "View System Resources": "htop"
    },
    "extensions": {
        "github": {
            "origin": "~/Developer/github.com/pomdtr/sunbeam/extensions/github.sh",
            // preferences for the extension, use it to pass config or secrets
            "preferences": {
                "token": "xxxx"
            },
            // additional root items to show
            "items": [
                {
                    "title": "List Sunbeam Issues",
                    "command": "list-issues",
                    "params": {
                        "repo": "pomdtr/sunbeam",
                    }
                }
            ]
        }
    }
}
EOF
```
