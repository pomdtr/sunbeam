# Payload

The payload is passed as the first argument to the script when a command is run.

```json
{
    // the command name
    "command": "say-hello",
    // if the command defines parameters, they are passed here
    "params": {
        "name": "Steve"
    },
    // if the manifest defines preferences, they are passed here
    "preferences": {
        "show-hidden-files": true
    },
    // the current working directory of the user
    "cwd": "/home/steve",
    // only set if the command is a search
    "query": "Hello, Steve!"
}
```
