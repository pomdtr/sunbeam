#!/bin/bash

set -x
set -euo pipefail

./docs/examples/tldr/sunbeam-command 2>/dev/null | sunbeam validate
./docs/examples/devdocs/sunbeam-command 2>/dev/null | sunbeam validate
./docs/examples/file-browser/sunbeam-command 2>/dev/null | sunbeam validate
./docs/getting-started/dynamic-list.py 2>/dev/null | sunbeam validate
./docs/getting-started/dynamic-list-with-args.py 2>/dev/null | sunbeam validate
cat ./docs/getting-started/static-list.json 2>/dev/null | sunbeam validate
