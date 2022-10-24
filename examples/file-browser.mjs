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
    return {
      title: file.name,
      subtitle: filepath,
      actions: [
        lstat.isDirectory()
          ? {
              type: "push",
              title: "Browse Directory",
              target: "sunbeam/file-browser",
              params: {
                root: filepath,
              },
            }
          : { type: "open", title: "Open File", path: filepath },
      ],
    };
  })
);

for (const item of items) {
  console.log(JSON.stringify(item));
}
