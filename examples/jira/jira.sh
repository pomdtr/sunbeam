#!/bin/bash

JQL=$1

sunbeam http \
    --auth "achille.lacoin@dailymotion.com:$JIRA_TOKEN" \
    --form jql="$JQL" \
    "https://dailymotion.atlassian.net/rest/api/2/search" |
sunbeam query '.issues[] | {
    title: .fields.summary,
    subtitle: .key,
    actions: [
        {
            type: "open",
            target: "https://dailymotion.atlassian.net/browse/\(.key)"
        }
    ],
    accessories: [
        "\(.fields.status.name)"
    ]
}' | sunbeam query --slurp '{
    type: "list",
    list: {items: .}
}'

