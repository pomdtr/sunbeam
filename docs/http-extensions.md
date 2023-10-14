# HTTP Extensions

In addition to local extensions, sunbeam can also interact with extensions hosted on a remote server.

The set of constraints that apply to local extensions also apply to remote extensions. Please refer to the [script extensions](./script-extensions.md) section for more details.

- A get request to the extension url must return a valid extension manifest.

    ```http
    GET https://remote-extension
    {
        "title": "Remote Extension",
        "commands": [{
            "name": "remote-command",
            "title": "Remote Command",
            "mode": "view"
        }]
    }
    ```

- A POST request to the extension url must return a valid extension manifest.

    ```http
    POST https://my-extension/remote-command
    {
        "type": "list",
        "items": [...]
    }
    ```

You can install a remote extension using the `sunbeam extension install <url>` command.

Feel free to add a sunbeam endpoint to your own server, and share your extensions with the community!

## Examples

Val Town is an excellent service to host sunbeam extensions. Here is an valtown extension for sunbeam, implemented in valtown as a remote extension:

<iframe src="https://www.val.town/embed/pomdtr.sunbeamValTownFn" width="100%" height="500" frameBorder="0" />
