# Input

## Common Fields

```json
{
    "label": "Label", // required
    "name": "name", // required
    "type": "text", // required
    "optional": true // optional
}
```

## Text

```json
{
    "type": "text", // required
    "name": "message", // required
    "label": "title", // required
    "optional": true,
    "text": {
        "placeholder": "Enter your title",
        "default": "Hi",
    }
}
```

## Password

```json
{
    "type": "password", // required
    "name": "password", // required
    "label": "Password", // required
    "optional": true,
    "password": {
        "placeholder": "Enter your password",
    }
}
```

## Textarea

```json
{
    "type": "textarea", // required
    "name": "message", // required
    "label": "Message", // required
    "optional": true,
    "textarea": {
        "placeholder": "Enter your message",
        "default": "Hello World"
    }
}
```

## Number

```json
{
    "type": "number", // required
    "name": "age", // required
    "label": "Age", // required
    "optional": true,
    "number": {
        "placeholder": "Enter your age",
        "default": 18,
    }
}
```

## Checkbox

```json
{
    "type": "checkbox", // required
    "name": "hidden", // required
    "label": "Show hidden entries", // required
    "checkbox": {
        "default": false
    }
}
```
