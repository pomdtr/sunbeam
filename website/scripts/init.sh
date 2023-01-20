#!/bin/bash
set -e

DIRNAME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DOCS_DIR="$DIRNAME/../src"
ROOT_DIR="$DIRNAME/../.."

function build_cobra_pages() {
	SED="sed"
	if which gsed >/dev/null 2>&1; then
		SED="gsed"
	fi

	CMD_DIR="$DOCS_DIR/cmd"
	rm "$CMD_DIR"/*.md
	DISABLE_EXTENSIONS=1 go run "$ROOT_DIR" docs "$CMD_DIR"
	"$SED" \
		-i'' \
		-e 's/SEE ALSO/See also/g' \
		-e 's/^## /# /g' \
		-e 's/^### /## /g' \
		-e 's/^#### /### /g' \
		-e 's/^##### /#### /g' \
		"$CMD_DIR"/*.md
}

function copy_json_schemas() {
	SCHEMAS_DIR="$DOCS_DIR/public/schemas"
	rm -rf "$SCHEMAS_DIR"
	mkdir -p "$SCHEMAS_DIR"
	cp -r "$ROOT_DIR"/app/schemas/* "$SCHEMAS_DIR"
}

build_cobra_pages
copy_json_schemas
