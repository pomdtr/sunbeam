# Inputs

## Textfield

```jsonc
{
    "type": "textfield", // required
    "title": "Name", // required
}
```

## Textarea

```jsonc
{
    "type": "textarea", // required
    "title": "Message", // required
}
```

## Password

```jsonc
{
    "type": "password", // required
    "title": "Password", // required
}
```

## Checkbox

```jsonc
{
    "type": "checkbox", // required
    "title": "Remember me", // optional
    "label": "When checked, you will be remembered", // optional
}
```

## Dropdown

```jsonc
{
    "type": "dropdown", // required
    "title": "Country", // required
    "choices": [ // required
       "United States",
       "Canada",
       "Mexico"
    ]
}
```

## File Browser

```jsonc
{
    "type": "file-browser", // required
    "title": "File", // required
    "fileTypes": [ // optional
        "png",
        "jpg",
        "gif"
    ]
}
```
