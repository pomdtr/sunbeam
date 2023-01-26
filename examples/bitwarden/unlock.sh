#!/bin/bash

set -euo pipefail

SESSION_TOKEN=$(bw unlock "$1" --raw)
sunbeam kv set session "$SESSION_TOKEN"
