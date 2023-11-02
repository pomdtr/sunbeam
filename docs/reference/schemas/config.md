# Config

The config is loaded from the `~/.config/sunbeam/config.json`, or `XDG_CONFIG_HOME/sunbeam/config.json` if set.

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
    // env variables loaded by sunbeam before running extensions
    "env": {
        "GITHUB_TOKEN": "ghp_xxx",
        "RAINDROP_TOKEN": "xxx"
    },
    // load env variables from a file
    "envFile": "secrets.env",
}
```
