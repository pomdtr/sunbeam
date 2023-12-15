# Input

## Text

```json
{
    "type": "text", // required
    "name": "message", // required
    "title": "title", // required
    "default": "Hi",
    "optional": true,
    "placeholder": "Enter your title",
}
```

## Password

```json
{
    "type": "password", // required
    "name": "password", // required
    "title": "Password", // required
    "optional": true,
    "placeholder": "Enter your password",
}
```

## Textarea

```json
{
    "type": "textarea", // required
    "name": "message", // required
    "title": "Message", // required
    "default": "Hello World",
    "optional": true,
    "placeholder": "Enter your message",
}
```

## Number

```json
{
    "type": "number", // required
    "name": "age", // required
    "title": "Age", // required
    "default": 18,
    "optional": true,
    "placeholder": "Enter your age",
}
```

## Checkbox

```json
{
    "type": "checkbox", // required
    "name": "hidden", // required
    "title": "Hidden",
    "label": "Show hidden entries", // required
    "default": false
}
```
