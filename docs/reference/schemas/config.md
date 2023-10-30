# Config

The config is loaded from the `~/.config/sunbeam/config.json`, or `XDG_CONFIG_HOME/sunbeam/config.json` if set.

```json
{
    // additional items to show in the root list
    "root": [
        {
            "title": "Search Overreact Feed",
            "extension": "rss",
            "command": "show",
            "params": {
                "url": "https://overreacted.io/rss.xml"
            }
        },
    ],
    // env variables loaded by sunbeam before running extensions
    "env": {
        "GITHUB_TOKEN": "ghp_xxx",
        "RAINDROP_TOKEN": "xxx"
    }
}
```
