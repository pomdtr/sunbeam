#!/usr/bin/env -S deno run -A
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as path from "https://deno.land/std/path/mod.ts";

const dirname = new URL(".", import.meta.url).pathname;

const extensions: Record<string, sunbeam.ExtensionConfig> = {};
const entries = Deno.readDirSync(path.join(dirname, "..", "extensions"));
for (const entry of entries) {
  const extension = path.extname(entry.name);
  const stem = path.basename(entry.name, extension);

  extensions[stem] = {
    origin: `./extensions/${entry.name}`,
  };
}

const config: sunbeam.Config = {
  $schema: "./internal/schemas/config.schema.json",
  extensions,
};

Deno.writeTextFileSync(
  path.join(dirname, "..", "sunbeam.json"),
  JSON.stringify(config, null, 2)
);
