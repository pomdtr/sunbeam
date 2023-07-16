#!/usr/bin/env bash

# @title {{ .Title }}
# @description {{ .Description }}


if [ $# -eq 0 ]; then
    payload='{"args": []}'
else
    # shellcheck disable=SC2016
    payload=$(printf '%s\n' "$@" | sunbeam query -R '. as $input | try fromjson catch $input' | sunbeam query -s '{args: .}')
fi

url="https://api.val.town/v1/run/{{ .Val }}"
if [ -n "$VALTOWN_TOKEN" ]; then
    sunbeam fetch "$url" -H 'Content-Type: application/json' -H "Authorization: Bearer $VALTOWN_TOKEN" -d "$payload"
else
    sunbeam fetch "$url" -H 'Content-Type: application/json' -d "$payload"
fi
