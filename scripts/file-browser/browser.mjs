#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Browse Directory
// @sunbeam.packageName File Browser
// @sunbeam.mode interactive
// @sunbeam.argument1 { "type": "text", "placeholder": "path" }

import * as path from "path";
import * as fs from "fs/promises";
import * as os from "os";
const { params } = JSON.parse(await stdin());

let root;
if (params && params.root) {
  root = params.root;
} else {
  root = argv._[0];
}

let files = await fs.readdir(root, { withFileTypes: true });
files = files.filter((file) => !file.name.startsWith("."));

const items = await Promise.all(
  files.map(async (file) => {
    const filepath = path.join(root, file.name);
    const lstat = await fs.lstat(filepath);
    return {
      title: file.name,
      subtitle: filepath,
      actions: [
        lstat.isDirectory()
          ? {
              type: "callback",
              title: "Browse Directory",
              push: true,
              params: {
                root: filepath,
              },
            }
          : { type: "open", title: "Open File", path: filepath },
      ],
    };
  })
);

items.push({
  title: "..",
  subtitle: path.dirname(root),
  actions: [
    {
      type: "callback",
      title: "Go Up",
      push: true,
      params: { root: path.dirname(root) },
    },
  ],
});

console.log(
  JSON.stringify({
    type: "list",
    list: {
      items,
    },
  })
);
