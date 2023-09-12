#!/usr/bin/env -S deno run --allow-read

import * as sunbeam from "../../../pkg/typescript/mod.ts";
import * as path from "https://deno.land/std@0.201.0/path/mod.ts";
import { handler } from "../../../pkg/typescript/http.ts";

const extension = new sunbeam.Extension({
  title: "File Browser",
  description: "A file browser extension for Sunbeam",
}).command({
  name: "browse",
  title: "Browse files",
  mode: "filter",
  params: [
    {
      name: "root",
      type: "string",
      optional: true,
    },
  ],
  run: async ({ params }) => {
    const root = params.root as string || Deno.cwd();

    const entries = [];
    for await (const entry of Deno.readDir(root)) {
      entries.push(entry);
    }

    const items: sunbeam.Listitem[] = entries.map((entry) => {
      const actions: sunbeam.Action[] = [];
      const filepath = path.join(root, entry.name);

      if (entry.isDirectory) {
        actions.push({
          title: "Browse Directory",
          type: "run",
          command: {
            name: "browse",
            params: {
              root: path.join(root, entry.name),
            },
          },
        });
      }

      return {
        title: entry.name,
        subtitle: filepath,
        actions,
      };
    });

    return {
      title: `${root}`,
      items,
    };
  },
});

Deno.serve(handler(extension));
