#!/usr/bin/env -S deno run -A
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest: sunbeam.Manifest = {
  title: "NPM Search",
  description: "Search NPM packages",
  commands: [
    {
      name: "search",
      title: "Search NPM Packages",
      mode: "search",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (!payload.query) {
  const list = { emptyText: "Enter a search query" };
  console.log(JSON.stringify(list));
  Deno.exit(0);
}

const resp = await fetch(
  `https://registry.npmjs.com/-/v1/search?text=${
    encodeURIComponent(
      payload.query,
    )
  }`,
);
const { objects: packages } = await resp.json();
const items: sunbeam.ListItem[] = [];
for (const pkg of packages) {
  const item: sunbeam.ListItem = {
    title: pkg.package.name,
    subtitle: pkg.package.description || "",
    actions: [
      {
        type: "open",
        title: "Open Package",
        url: pkg.package.links.npm,
      },
      {
        type: "copy",
        title: "Open Package Name",
        text: pkg.package.name,
        exit: true,
      },
    ],
  };

  items.push(item);
}
console.log(JSON.stringify({ items }));
