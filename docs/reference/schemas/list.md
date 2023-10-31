# List

```json
{
    // the list of items to display (required)
    "items": [
        {
            // title of the item (required)
            "title": "sunbeam",
            // subtitle of the item (optional)
            // will be displayed at the right of the title, in a faint color
            "subtitle": "pomdtr",
            // the list of accessories (optional)
            // they will be displayed on the right side of the item
            "accessories": [
                "225 *",
                "public"
            ],
            // unique identifier of the item (optional)
            // if not set, the title will be used as id
            "id": "pomdtr/sunbeam",
            // the list of actions that can be performed on the item (optional)
            "actions": [
                {
                    "title": "Open in Browser",
                    "type": "open",
                    "target": "https://github.com/pomdtr/sunbeam"
                }
            ]
        }
    ],
    // whether the list is dynamic or not (optional)
    // if true, the list will be refreshed every time the user types a character
    "dynamic": false,
    // the text to display when the list is empty (optional)
    "emptyText": "No items found",
    // the list of actions shown when no item is selected (optional)
    "actions": [
        {
            "title": "Refresh Items",
            "type": "reload"
        }
    ]
}
```
