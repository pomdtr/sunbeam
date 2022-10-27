#!/usr/bin/env zx

$.verbose = false;

const { stdout } =
  await $`kubectl get namespace --output go-template='{{ range .items }}{{.metadata.name}}{{"\\n"}}{{ end }}'`;

const namespaces = stdout.split("\n").filter((row) => row);

const items = namespaces.map((namespace) => ({
  title: namespace,
  actions: [
    {
      title: "Switch to",
      type: "exec",
      command: `kubectl config set-context --current --namespace=${namespace}`,
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
