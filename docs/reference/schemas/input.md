# Input

## Text

```json
{
    // the type of the input (required)
    "type": "text",
    // the title of the input (required)
    "title": "Name",
    // whether the input is required (optional)
    "required": true,
    // the placeholder of the input (optional)
    "placeholder": "Your name",
    // the default value of the input (optional)
    "default": "Steve Jobs"
}
```

## TextArea

```json
{
    // the type of the input (required)
    "type": "textarea",
    // the title of the input (required)
    "title": "Query",
    // whether the input is required (optional)
    "required": true,
    // the placeholder of the input (optional)
    "placeholder": "Query",
    // the default value of the input (optional)
    "default": "SELECT * FROM users"
}
```

## Checkbox

```json
{
    // the type of the input (required)
    "type": "checkbox",
    // the title of the input (required)
    "title": "Admin",
    // whether the input is required (optional)
    "required": true,
    // the default value of the input (required)
    "label": "Is admin?",
    // the default value of the input (optional)
    "default": false
}
```

## Select

```json
{
    // the type of the input (required)
    "type": "select",
    // the title of the input (required)
    "title": "Language",
    // whether the input is required (optional)
    "required": true,
    // the options of the input (required)
    "options": [
        {
            "title": "Go",
            // the value of the option (required)
            // the value can of type string, number or boolean
            "value": "go"
        },
        {
            "title": "Rust",
            "value": "rust"
        },
        {
            "title": "Deno",
            "value": "deno"
        }
    ],
    // the default value of the input (optional)
    "default": "go"
}
```
