#!/usr/bin/env -S deno run -A

import { DB } from "https://deno.land/x/sqlite@v3.8/mod.ts";
import * as fs from "https://deno.land/std@0.203.0/fs/mod.ts";
import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as path from "https://deno.land/std@0.186.0/path/mod.ts";

const manifest = {
  title: "VS Code",
  description: "Manage your VS Code projects",
  root: [
    {
      command: "list-projects",
    },
  ],
  commands: [
    {
      name: "list-projects",
      title: "List Projects",
      mode: "filter",
    },
    {
      name: "open-project",
      title: "Open Project",
      mode: "silent",
      params: [
        { name: "path", title: "Project Path", type: "string" },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);

if (payload.command == "list-projects") {
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

  const items: sunbeam.ListItem[] = entries.map((entry) => {
    const folderUri = new URL(entry.folderUri);
    const folderPath = path.dirname(folderUri.pathname);

    return {
      title: path.basename(folderUri.pathname),
      subtitle: path.basename(folderPath),
      actions: [
        {
          title: "Open in VS Code",
          command: "open-project",
          params: {
            path: folderPath,
          },
        },
        {
          title: "Open Folder",
          extension: "std",
          command: "open",
          params: {
            url: entry.folderUri,
          },
        },
        {
          title: "Copy Path",
          extension: "std",
          command: "copy",
          params: {
            text: entry.folderUri.replace("file://", ""),
          },
        },
      ],
    };
  });

  const list: sunbeam.List = { items };
  console.log(JSON.stringify(list));
} else if (payload.command == "open-project") {
  const { path } = payload.params;
  await new Deno.Command("code", {
    args: [path],
  }).output();
}
