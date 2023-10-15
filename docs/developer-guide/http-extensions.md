# HTTP Extensions

In addition to local extensions, sunbeam can also interact with extensions hosted on a remote server.

The experience of developing an http extension is similar to that of developing a script extension. The main difference is that the extension is hosted on a remote server, and is accessed using http requests.

## Constraints

A GET request to the extension root must return a valid extension manifest.

```
GET https://github-extension.io
Accept: application/json

{
    "title": "Github Extension",
    "commands": [{
        "name": "list-issues",
        "title": "Remote Command",
        "mode": "view",
        "params": [
            {
                "name": "full_name",
                "type": "string",
                "required": true,
                "description": "full name of the repository",
            }
        ]
    }]
}
```

POST request are used to execute commands.

```
POST https://github-extension.io/list-issues
Content-Type: application/json
Accept: application/json

{
    "params": {
        "full_name": "pomdtr/sunbeam"
    }
}

{
    "type": "list",
    "title": "Issues",
    "items": [...]
}
```

## Installation

You can install a remote extension using the `sunbeam extension install <url>` command.

Feel free to add a sunbeam endpoint to your own server, and share your extensions with the community!

## Converting a Script Extension to an HTTP Extension

Just use the `sunbeam serve` command to serve your script extension as an HTTP extension.

```
sunbeam serve ./sunbeam-devdocs --port 8080
```

## Recommended Services to Host HTTP Extensions

### Val Town (TypeScript)

[Val Town](https://www.val.town/) is a social website to write and deploy TypeScript.

Hosting sunbeam extensions on Val Town is free and easy. An example extension is available here: [pomdtr.sunbeamValTownFn](https://www.val.town/v/pomdtr.sunbeamValTownFn).

### CodeSandbox

[CodeSandbox](https://codesandbox.io) is an online editor that helps you create web applications. It supports any language and framework thanks to Docker.

Each sandox gets an http endpoint that can be used to host sunbeam extensions.


### Deno Deploy (TypeScript)

If you want to author your extension in TypeScript, but prefer to work from your own editor, [Deno Deploy](https://deno.com/deploy) is a good alternative to Val Town.


### Deta Space

[Deta](https://deta.space) is a personal cloud that combines a database, a file storage, and a serverless platform.

