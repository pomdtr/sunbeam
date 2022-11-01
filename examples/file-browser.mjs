#!/usr/bin/env zx

import * as path from "path";
import * as os from "os";
import * as fs from "fs/promises";

const currentDir = argv._[0];
const showHidden = argv.all;

let files = await fs.readdir(currentDir, { withFileTypes: true });
if (!showHidden) {
  files = files.filter((file) => !file.name.startsWith("."));
}

const items = await Promise.all(
  files.map(async (file) => {
    const filepath = path.join(currentDir, file.name);
    const prettyPath = filepath.replace(os.homedir(), "~");
    const lstat = await fs.lstat(filepath);
    const primaryAction = lstat.isDirectory()
      ? {
          type: "run",
          title: "Browse Directory",
          shortcut: "enter",
          target: "browse-files",
          params: {
            root: filepath,
            all: !!showHidden,
          },
        }
      : {
          type: "open-file",
          title: "Open File",
          shortcut: "enter",
          path: filepath,
        };
    return {
      title: `${file.name}`,
      subtitle: prettyPath,
      actions: [
        primaryAction,
        {
          type: "reload",
          title: "Toggle Hidden Files",
          shortcut: "ctrl+h",
          params: {
            all: !showHidden,
          },
        },
      ],
    };
  })
);

for (const item of items) {
  console.log(JSON.stringify(item));
}
