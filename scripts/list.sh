#!/bin/sh

CAT << EOF | sunbeam -stdin
{
    "type": "list",
    "list": {
        "items": [
            {"title": "A"},
            {"title": "A"},
            {"title": "A"}
        ]
    }
}
EOF
