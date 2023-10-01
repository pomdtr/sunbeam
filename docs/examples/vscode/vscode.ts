#!/usr/bin/env -S deno run --allow-env --allow-read --allow-write --allow-run

import { DB } from "https://deno.land/x/sqlite@v3.8/mod.ts";
import * as fs from "https://deno.land/std@0.203.0/fs/mod.ts";

if (Deno.args.length === 0) {
  console.log(JSON.stringify({
    title: "VS Code",
    root: "list-projects",
    commands: [
      {
        name: "list-projects",
        title: "List Projects",
        mode: "view",
      },
    ],
  }));
  Deno.exit(0);
}

const homedir = Deno.env.get("HOME");
const db = new DB(
  `${homedir}/Library/Application Support/Code/User/globalStorage/state.vscdb`,
);
const res = db.query(
  "SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'",
);

// deno-lint-ignore no-explicit-any
let entries: any[] = JSON.parse(res[0][0] as string);
entries = entries.filter((entry) => {
  if (!entry.folderUri) {
    return false;
  }

  const path = entry.folderUri.replace("file://", "");
  if (!fs.existsSync(path)) {
    return false;
  }

  return true;
});

const items = entries.map((entry) => ({
  title: entry.folderUri.split("/").pop(),
  accessories: [
    entry.folderUri.replace("file://", "").replace(homedir, "~"),
  ],
  actions: [
    {
      title: "Open in VS Code",
      onAction: {
        type: "open",
        target: entry.folderUri,
        app: {
          name: "Visual Studio Code",
          platform: "mac",
        },
        exit: true,
      },
    },
    {
      title: "Open Folder",
      key: "o",
      onAction: { type: "open", target: entry.folderUri, exit: true },
    },
    {
      title: "Copy Path",
      key: "c",
      onAction: {
        type: "copy",
        exit: true,
        text: entry.folderUri.replace("file://", ""),
      },
    },
  ],
}));

console.log(
  JSON.stringify({
    type: "list",
    items,
  }),
);
