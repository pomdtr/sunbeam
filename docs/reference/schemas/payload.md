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
    // the current working directory of the user
    "cwd": "/home/steve",
    // the user query, only accessible in dynamic lists
    "query": "Hello, Steve!"
}
```
