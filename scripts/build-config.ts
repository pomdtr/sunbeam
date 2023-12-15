#!/usr/bin/env -S deno run -A
import { Config, ExtensionConfig } from "../sdk/config.ts";
import * as path from "https://deno.land/std@0.209.0/path/mod.ts";

const dirname = new URL(".", import.meta.url).pathname;

const extensions: Record<string, ExtensionConfig> = {};
const entries = Deno.readDirSync(path.join(dirname, "..", "extensions"));
for (const entry of entries) {
  const extension = path.extname(entry.name);
  const stem = path.basename(entry.name, extension);

  extensions[stem] = {
    origin: `./extensions/${entry.name}`,
  };
}

const config: Config = {
  $schema: "./internal/schemas/config.schema.json",
  oneliners: [
    {
      title: "Refresh Config",
      command: "./scripts/build-config.ts",
    },
  ],
  extensions,
};

Deno.writeTextFileSync(
  path.join(dirname, "..", "sunbeam.json"),
  JSON.stringify(config, null, 2) + "\n"
);
