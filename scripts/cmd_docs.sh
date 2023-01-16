#!/bin/bash
set -e

SED="sed"
if which gsed >/dev/null 2>&1; then
	SED="gsed"
fi

targetDir="website/cmd"
rm "$targetDir"/*.md
DISABLE_EXTENSIONS=1 go run . docs "$targetDir"
"$SED" \
	-i'' \
	-e 's/SEE ALSO/See also/g' \
	-e 's/^## /# /g' \
	-e 's/^### /## /g' \
	-e 's/^#### /### /g' \
	-e 's/^##### /#### /g' \
	"$targetDir"/*.md
