# Manifest

The manifest is a JSON file that describes the extension.
It contains the title, description and the list of commands provided by the extension.

```json
{
  // the title of the extension, will be shown in the root list
  "title": "DevDocs",
  // the description of the extension, will be shown in usage string
  "description": "Search DevDocs.io",
  // see input schema
  "preferences": [
    {
      "name": "hidden",
      "title": "Show hidden entries",
      "type": "boolean"
    }
  ],
  "commands": [
    {
      // unique identifier of the command (required)
      "name": "list-entries",
      // the title of the command, will be shown in the root list (required)
      "title": "List Entries from Docset",
      // the mode of the command, can be "filter", "search", "detail", "tty", "silent" (required)
      // if you want to display a list of items that can be filtered, use the filter mode
      // if you want to refresh the list of items every time the user types a character, use the search mode
      // if you want to display a static view, use the view mode
      // use the tty mode if you want to use the terminal directly
      // or use the silent mode if you don't want to display anything
      "mode": "filter",
      // whether the command should be hidden from the root list (optional)
      "hidden": false,
      // the list of parameters for the command (optional)
      // see the input schema for more details
      "params": [
        {
          "name": "slug",
          "type": "string", // can be "string", "number", "boolean"
          "title": "Docset Slug",
        }
      ]
    }
  ]
}
```
