# Manifest

The manifest is a JSON file that describes the extension.
It contains the title, description and the list of commands provided by the extension.

```json
{
  // the title of the extension, will be shown in the root list
  "title": "DevDocs",
  // the description of the extension, will be shown in usage string
  "description": "Search DevDocs.io",
  // additional items to show in the root list (optional)
  "root": [
    {
      "title": "Search Deno Docs",
      "command": "list-entries",
      "params": {
        "slug": "deno"
      }
    }
  ],
  // the list of commands required by the extension (optional)
  "requirements": {
    "name": "jq",
    "link": "https://stedolan.github.io/jq/"
  },
  "commands": [
    {
      // unique identifier of the command (required)
      "name": "list-entries",
      // the title of the command, will be shown in the root list (required)
      "title": "List Entries from Docset",
      // the mode of the command, can be "list", "detail", "tty", "silent" (required)
      // if you want to display a list of items, use the list mode
      // if you want to display a detail view, use the detail mode
      // use the tty mode if you want to use the terminal directly
      // or use the silent mode if you don't want to display anything
      "mode": "page",
      // the list of parameters for the command (optional)
      "params": [
        {
          // unique identifier of the parameter (required)
          "name": "slug",
          // type of the parameter (required)
          // can be "string", "number", or "boolean"
          "type": "string",
          // whether the parameter is required or not (default: false)
          // if a a command has a required parameter, it will not be shown in the root list
          "required": false,
          // description of the parameter (optional)
          // will be shown in the usage string
          "description": "docset to search",
          // default value of the parameter (optional)
          // the type of the default value must match the type of the parameter
          "default": "go"
        }
      ]
    }
  ]
}
```
