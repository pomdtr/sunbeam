#!/bin/bash

JQL=$1

curl -X GET \
    -G --data-urlencode jql="$JQL" \
    -u "achille.lacoin@dailymotion.com:$JIRA_TOKEN" \
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
    items: .
}'

