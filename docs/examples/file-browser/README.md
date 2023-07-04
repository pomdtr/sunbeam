# File Browser (Python)

## Requirements

- [python](https://www.python.org/)

## Demo

![demo](./demo.gif)

## Install

```bash
sunbeam extension install files https://raw.githubusercontent.com/pomdtr/sunbeam/main/docs/examples/file-browser/sunbeam-command
```

## Usage

```bash
sunbeam files # Browse the current directory
sunbeam files /path/to/directory # Browse a specific directory
sunbeam files --show-hidden # Show hidden files
```

## Code

```python
{{#include ./sunbeam-command}}
```
