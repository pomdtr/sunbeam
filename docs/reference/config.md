# Config

The config will be searched in the following locations:

- `SUNBEAM_CONFIG` environment variable if set
- `XDG_CONFIG_HOME/sunbeam/config.json` if `XDG_CONFIG_HOME` is set
- `$HOME/.config/sunbeam/config.json`

If no config is found, sunbeam will create one.

```json
{
    // additional items to show in the root list
    "root": [
        {
            // title of the item shown in the list
            "title": "Search Overreact Feed",
            // command to run when the item is chosen
            "command": "sunbeam rss show --url https://overreacted.io/rss.xml"
        },
        {
            "title": "View System Resources",
            "command": "htop"
        }
    ],
    // load env variables from a file
    // relative paths are resolved from the config file location
    "envFile": "secrets.env",
}
```
