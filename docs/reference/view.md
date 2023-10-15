# View

## List

```json
{
    // the type of the view (required)
    "type": "list",
    // the title of the view (optional)
    "title": "Github Repositories",
    // whether the list is dynamic or not (optional)
    // if true, the list will be refreshed every time the user types a character
    "dynamic": false,
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
                    // a command to execute when the action is triggered (required)
                    // see the command section for more details
                    "onAction": {
                        "type": "open",
                        "target": "https://github.com/pomdtr/sunbeam"
                    }
                }
            ]
        }
    ]
}
```

## Detail

```json
{
    // the type of the view (required)
    "type": "detail",
    // the title of the view (optional)
    "title": "Sunbeam Readme",
    // the text to display (required)
    "markdown": "# Sunbeam\n\n***the love child of raycast and fzf***",
    // the list of actions that can be performed on the view (optional)
    "actions": [
        {
            "title": "Open Sunbeam Website",
            "onAction": {
                "type": "open"
                "target": "https://pomdtr.github.io/sunbeam"
            }
        }
    ]
}
```

## Form

```json
{
    // the type of the view (required)
    "type": "form",
    // the title of the view (optional)
    "title": "Create Repository",
    // the fields of the form (required)
    "fields": [
        {
            // the title of the field (required)
            "title": "Name",
            // an unique identifier for the field (required)
            "name": "name",
            "input": {
                "type": "text",
                "placeholder": "e.g. sunbeam",
            }
        },
        {
            "title": "Description",
            "name": "description",
            "input": {
                "type": "textarea",
                "placeholder": "e.g. the love child of raycast and fzf",
            }
        },
        {
            "title": "Private",
            "name": "checkbox",
            "input": {
                "type": "checkbox",
                "label": "Whether the repository should be private or not",
            }
        },
        {
            "title": "License",
            "name": "license",
            "input": {
                "type": "select",
                "choices": [
                    { "title": "MIT", "value": "mit" },
                    { "title": "Apache", "value": "apache" },
                    { "title": "GPL", "value": "gpl" }
                ],
                "default": "mit"
            }
        }
    ]
}
```
