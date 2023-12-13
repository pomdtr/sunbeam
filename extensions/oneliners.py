#!/usr/bin/env python3

import json
import sys
import os
import subprocess

if len(sys.argv) == 1:
    manifest = {
        "title": "Oneliners",
        "description": "Manage your oneliners",
        "commands": [
            {
                "name": "create",
                "title": "Create Oneliner",
                "mode": "silent",
                "params": [
                    {
                        "name": "title",
                        "title": "Title",
                        "type": "text",
                    },
                    {
                        "name": "command",
                        "title": "Command",
                        "type": "text",
                    },
                    {
                        "name": "exit",
                        "title": "Exit",
                        "type": "checkbox",
                        "label": "Exit after running command",
                        "optional": True
                    }
                ]
            },
            {
                "name": "delete",
                "title": "Delete Oneliner",
                "hidden": True,
                "mode": "silent",
                "params": [
                    {
                        "name": "index",
                        "title": "Index",
                        "type": "number",
                    },
                ]
            },
            {
                "name": "edit",
                "title": "Edit Oneliner",
                "hidden": True,
                "mode": "silent",
                "params": [
                    {
                        "name": "index",
                        "title": "Index",
                        "type": "number",
                    },
                    {
                        "name": "title",
                        "title": "Title",
                        "type": "text",
                    },
                    {
                        "name": "command",
                        "title": "Command",
                        "type": "text",
                    },
                    {
                        "name": "exit",
                        "title": "Exit",
                        "type": "checkbox",
                        "label": "Exit after running command",
                    }
                ]
            },
            {
                "name": "manage",
                "title": "Manage Oneliners",
                "mode": "filter"
            },
            {
                "name": "run",
                "title": "Run Oneliner",
                "key": "r",
                "hidden": True,
                "mode": "tty",
                "params": [
                    {
                        "name": "index",
                        "title": "Index",
                        "type": "number",
                    },
                ]
            }
        ]
    }
    print(json.dumps(manifest))
    sys.exit(0)

if sunbeam_config := os.environ.get("SUNBEAM_CONFIG"):
    config_path = sunbeam_config
elif config_home := os.environ.get("XDG_CONFIG_HOME"):
    config_path = os.path.join(config_home, "sunbeam", "sunbeam.json")
else:
    config_path = os.path.join(os.path.expanduser("~"), ".config", "sunbeam", "sunbeam.json")

with open(config_path) as f:
    config = json.load(f)

payload = json.loads(sys.argv[1])
if payload['command'] == 'add':
    config.setdefault('oneliners', [])
    config['oneliners'].append(payload['params'])

    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)
if payload['command'] == 'edit':
    idx = payload['params'].pop('index')

    config['oneliners'][idx] = payload['params']
    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)
elif payload['command'] == 'run':
    idx = payload['params']['index']
    oneliner = config['oneliners'][idx]
    subprocess.run(["sh", "-c", oneliner['command']], stdout=sys.stdout, stderr=sys.stderr, stdin=sys.stdin)
elif payload['command'] == 'manage':
    items = []
    for idx, oneliner in enumerate(config.get('oneliners', [])):
        items.append({
            'title': oneliner['title'],
            'subtitle': oneliner['command'],
            'actions': [
                {
                    'title': "Edit Oneliner",
                    'key': 'e',
                    'type': 'run',
                    "command": "edit",
                    "reload": True,
                    "params": {
                        "index": idx,
                        "title": {
                            "default": oneliner['title']
                        },
                        "command": {
                            "default": oneliner['command']
                        },
                        "exit": {
                            "default": oneliner.get('exit', False)
                        }
                    }
                },
                {
                    'title': 'Run Command',
                    'type': 'run',
                    'command': 'run',
                    'params': {
                        'index': idx
                    },
                    'exit': oneliner.get('exit', False)
                },
                {
                    'title': 'Copy Command',
                    'type': 'copy',
                    'text': oneliner['command'],
                    'exit': True
                },
                {
                    'title': "Create Oneliner",
                    'key': 'n',
                    'type': 'run',
                    "command": "add",
                    "reload": True
                },
                {
                    'title': "Delete Oneliner",
                    'key': 'd',
                    'type': 'run',
                    "reload": True,
                    "command": "delete",
                    "params": {
                        "index": idx
                    }
                }
            ]
        })
    print(json.dumps({"items": items}))
elif payload['command'] == 'delete':
    idx = payload['params']['index']
    config['oneliners'].pop(idx)
    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)


