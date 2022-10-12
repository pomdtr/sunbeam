#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Browse Directory
// @sunbeam.packageName File Browser
// @sunbeam.mode interactive
// @sunbeam.argument1 { "type": "text", "placeholder": "path" }

import * as path from "path";
import * as fs from "fs/promises";

const currentDir = argv._[0];

let files = await fs.readdir(currentDir, { withFileTypes: true });
files = files.filter((file) => !file.name.startsWith("."));

const items = await Promise.all(
  files.map(async (file) => {
    const filepath = path.join(currentDir, file.name);
    const lstat = await fs.lstat(filepath);
    return {
      title: file.name,
      subtitle: filepath,
      actions: [
        lstat.isDirectory()
          ? {
              type: "push",
              path: "./browser.mjs",
              args: [filepath],
              title: "Browse Directory",
            }
          : { type: "open", title: "Open File", path: filepath },
      ],
    };
  })
);

console.log(
  JSON.stringify({
    type: "list",
    list: {
      items,
    },
  })
);
