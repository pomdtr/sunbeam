#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Browse My Computer
// @sunbeam.mode command
// @sunbeam.packageName File Browser

import * as path from "path";
import * as fs from "fs/promises";
import * as os from "os";

$.verbose = false;

const startDir = os.homedir();

const files = await fs.readdir(startDir, { withFileTypes: true });

const items = files.map((file) => ({
  title: file.name,
  subtitle: path.join(startDir, file.name),
  fill: path.join(startDir, file.name),
  actions: [{ type: "open", path: path.join(startDir, file.name) }],
}));

console.log(
  JSON.stringify({
    type: "list",
    list: {
      items,
    },
  })
);
