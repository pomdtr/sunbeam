#!/bin/sh

CAT << EOF | go run main.go -stdin | xargs echo
{
    "type": "list",
    "list": {
        "items": [
            {"title": "A", "actions": [{"type": "callback", "title": "Action", "params": "A"}]},
            {"title": "B", "actions": [{"type": "callback", "title": "Action", "params": "B"}]},
            {"title": "C", "actions": [{"type": "callback", "title": "Action", "params": "C"}]}
        ]
    }
}
EOF
