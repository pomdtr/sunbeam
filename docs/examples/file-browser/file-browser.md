# File Browser (Python)

## Requirements

You will need to have [Python](https://www.python.org/) installed.

## Usage

```bash
sunbeam run ./file-browser.py # Browse the current directory
sunbeam run ./file-browser.py /path/to/directory # Browse a specific directory
sunbeam run -- ./file-browser.py --show-hidden # Show hidden files
```

## Code

```python
{{#include ./file-browser.py}}
```
