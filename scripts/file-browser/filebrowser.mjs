#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Browse My Computer
// @sunbeam.packageName File Browser
// @sunbeam.mode interactive

// @sunbeam.argument1 { "type": "text", "placeholder": "root", "required": false }

import * as path from "path";
import * as fs from "fs/promises";
import * as os from "os";

$.verbose = false;

const root = argv.root || os.homedir();

const files = await fs.readdir(root, { withFileTypes: true });

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
              type: "push",
              path: "file-browser/filebrowser.mjs",
              options: {
                root: filepath,
              },
            }
          : { type: "open", path: filepath },
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
