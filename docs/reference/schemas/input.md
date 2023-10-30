# Input

## Text

```json
{
    "type": "text",
    "title": "Name",
    "required": true,
    "placeholder": "Your name",
    "default": "Steve Jobs"
}
```

## TextArea

```json
{
    "type": "textarea",
    "title": "Query",
    "required": true,
    "placeholder": "Query",
    "default": "SELECT * FROM users"
}
```

## Checkbox

```json
{
    "type": "checkbox",
    "title": "Admin",
    "required": true,
    "label": "Is admin?",
    "default": false
}
```

## Select

```json
{
    "type": "select",
    "title": "Language",
    "required": true,
    "options": [
        {
            "title": "Go",
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
    "default": "go"
}
```
