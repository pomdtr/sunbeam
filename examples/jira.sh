#!/bin/bash

JIRA_TOKEN=$1
JQL=$2

curl -X GET \
    -u "achille.lacoin@dailymotion.com:$JIRA_TOKEN" \
    "https://dailymotion.atlassian.net/rest/api/2/search?jql=$JQL" |
sunbeam jq '.issues[] | {
    title: .key,
    subtitle: .fields.summary,
    actions: [
        {
            type: "open-url",
            url: "https://dailymotion.atlassian.net/browse/\(.key)"
        }
    ]
}'

