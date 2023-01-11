#!/bin/bash
set -e

SED="sed"
if which gsed >/dev/null 2>&1; then
	SED="gsed"
fi

rm -rf www/docs/cmd/*.md
DISABLE_EXTENSIONS=1 go run . docs www/docs/cmd
"$SED" \
	-i'' \
	-e 's/SEE ALSO/See also/g' \
	-e 's/^## /# /g' \
	-e 's/^### /## /g' \
	-e 's/^#### /### /g' \
	-e 's/^##### /#### /g' \
	./www/docs/cmd/*.md


for f in www/docs/cmd/*.md; do
    printf "%s\n" "- $(sed 's/www\/docs\///g' <<< "$f")"
done
