#!/usr/bin/env bash

set -x
set -euo pipefail

sudo apt-get update && sudo apt-get install direnv
npm install -g zx
