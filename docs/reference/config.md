# Config

The config will be searched in the following locations:

- `SUNBEAM_CONFIG` environment variable if set
- `XDG_CONFIG_HOME/sunbeam/config.json` if `XDG_CONFIG_HOME` is set
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
