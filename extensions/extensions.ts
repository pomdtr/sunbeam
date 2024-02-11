#!/usr/bin/env -S deno run -A

import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as path from "https://deno.land/std@0.208.0/path/mod.ts";

const manifest = {
  title: "Extensions",
  description: "Manage your extensions",
  commands: [
    {
      name: "list-extensions",
      title: "List Extensions",
      mode: "filter",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
const homeDir = Deno.env.get("HOME");
if (!homeDir) {
  console.error("HOME environment variable is required");
  Deno.exit(1);
}

const configPath = path.join(homeDir, ".config", "sunbeam", "sunbeam.json");
const config: sunbeam.Config = JSON.parse(Deno.readTextFileSync(configPath));
switch (payload.command) {
  case "list-extensions": {
    const items: sunbeam.ListItem[] = Object.entries(config.extensions).map(
      ([name, { origin }]) => ({
        title: name,
        subtitle: origin,
        actions: [{
          title: "Edit Extension",
          extension: "std",
          command: "edit",
          params: { path: origin },
          reload: true,
        }],
      } satisfies sunbeam.ListItem),
    );
    const list: sunbeam.List = {
      items,
    };

    console.log(JSON.stringify(list));
  }
}
