#!/usr/bin/env zx

import * as path from "path";
import * as fs from "fs/promises";

const currentDir = argv._[0];

let files = await fs.readdir(currentDir, { withFileTypes: true });
files = files.filter((file) => !file.name.startsWith("."));

const items = await Promise.all(
  files.map(async (file) => {
    const filepath = path.join(currentDir, file.name);
    const lstat = await fs.lstat(filepath);
    const primaryAction = lstat.isDirectory()
      ? {
          type: "launch",
          title: "Browse Directory",
          shortcut: "enter",
          target: "file-browser",
          params: {
            root: filepath,
          },
        }
      : {
          type: "open-file",
          title: "Open File",
          shortcut: "enter",
          path: filepath,
        };
    return {
      title: file.name,
      subtitle: filepath,
      actions: [primaryAction],
    };
  })
);

for (const item of items) {
  console.log(JSON.stringify(item));
}
