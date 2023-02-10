#!/bin/bash

set -euo pipefail

multipass list --format json | sunbeam query '.list[] |
{
    title: .name,
    subtitle: .release,
    preview: {
      command: "vm-info",
      with: {
        vm: .name
      }
    },
    accessories: [
      .state
    ],
    actions:
      (
        if
          .state == "Running"
        then
          [
            {type: "run-command", title: "Stop \(.name)", command: "stop-vm", onSuccess: "reload-page", with: {vm: .name}},
            {type: "run-command", title: "Attach", command: "open-shell", with: {vm: .name}}
          ]
        else
          [
            {type: "run-command", title: "Start \(.name)", command: "start-vm", onSuccess: "reload-page", with: {vm: .name}}
          ]
        end
      ),
}
' | sunbeam query --slurp '{
    type: "list",
    showPreview: true,
    items: .
}'
