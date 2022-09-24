#!/bin/sh

cat << EOF | go run main.go -stdin
{
    "type": "detail",
    "details": {
        "text": "This is a **markdown** string",
        "format": "markdown",
        "actions": [
            {"type": "copy", "title": "Action", "content": "This is the content to copy"}
        ]
    }
}
EOF
