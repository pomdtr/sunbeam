#!/bin/bash

set -x
set -euo pipefail

./docs/examples/tldr/tldr.sh 2>/dev/null | sunbeam validate
./docs/examples/devdocs/devdocs.sh 2>/dev/null | sunbeam validate
./docs/examples/file-browser/file-browser.py 2>/dev/null | sunbeam validate
./docs/getting-started/dynamic-list.py 2>/dev/null | sunbeam validate
./docs/getting-started/dynamic-list-with-args.py 2>/dev/null | sunbeam validate
cat ./docs/getting-started/static-list.json 2>/dev/null | sunbeam validate
