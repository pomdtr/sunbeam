# Manifest

The manifest is a JSON file that describes the extension.
It contains the title, description and the list of commands provided by the extension.

```json
{
  // the title of the extension, will be shown in the root list
  "title": "DevDocs",
  // the description of the extension, will be shown in usage string
  "description": "Search DevDocs.io",
  // items to show in the root list (optional)
  "items": [
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
  // see input schema
  "preferences": [
    {
      "name": "hidden",
      "label": "Show hidden entries",
      "type": "boolean",
      "default": false
    }
  ],
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
      "mode": "list",
      // the list of parameters for the command (optional)
      // see input schema
      "params": [
        {
          "name": "slug",
          "type": "text",
          "required": false,
          "title": "Docset Slug",
        }
      ]
    }
  ]
}
```
