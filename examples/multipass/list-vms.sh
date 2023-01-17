#!/bin/bash


multipass list --format json | sunbeam query '.list[] |
{
    title: .name,
    subtitle: .release,
    accessories: [
      .state
    ],
    actions:
      (
        if
          .state == "Running"
        then
          [
            {type: "run-command", title: "Stop \(.name)", onSuccess: "reload-page", command: "stop-vm", with: {vm: .name}},
            {type: "run-command", title: "Open Shell", command: "open-shell", with: {vm: .name}}
          ]
        else
          [
            {type: "run-command", title: "Start \(.name)", onSuccess: "reload-page", command: "start-vm", with: {vm: .name}}
          ]
        end
      ),
}
'
