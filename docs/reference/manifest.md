# Manifest

The manifest is a JSON file that describes the extension.
It contains the title, description and the list of commands provided by the extension.

```json
{
  // the title of the extension, will be shown in the root list
  "title": "DevDocs",
  // the description of the extension, will be shown in usage string
  "description": "Search DevDocs.io",
  // the list of commands provided by the extension
  "commands": [
    {
      // unique identifier of the command (required)
      "name": "list-entries",
      // the title of the command, will be shown in the root list (required)
      "title": "List Entries from Docset",
      // the mode of the command, can be "page", "silent" or "tty" (required)
      // if the mode is "page", the command is required to return a view on stdout when executed
      // if the mode is "silent" or "tty", the command can optionally return a command on stdout when executed
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
