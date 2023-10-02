# File Browser (Python)

## Requirements

- [python3](https://www.python.org)

## Install

```bash
curl -L https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/files/files.py > ~/.local/bin/sunbeam-files
chmod +x ~/.local/bin/sunbeam-files
```

## Usage

```bash
sunbeam files # browse current directory
sunbeam files ls --dir ~ --show-hidden # browse home directory and show hidden files
```

## Code

```bash
{{#include ./files.py}}
```
