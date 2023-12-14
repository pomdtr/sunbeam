#!/usr/bin/env -S deno run -A

import * as path from "https://deno.land/std@0.208.0/path/mod.ts";
const dirname = new URL(".", import.meta.url).pathname;
const rows = [];

rows.push(
  "---",
  "sidebar: false",
  "outline: 2",
  "---",
  "",
  "# Extension Catalog"
);

const extensionDir = path.join(dirname, "..", "extensions");
const entries = Deno.readDirSync(extensionDir);
for (const entry of entries) {
  const entrypoint = path.join(extensionDir, entry.name);
  const { stdout, success } = new Deno.Command(entrypoint).outputSync();
  if (!success) {
    console.error(`Failed to run entrypoint for ${entry.name}`);
    Deno.exit(1);
  }

  let manifest;
  try {
    manifest = JSON.parse(new TextDecoder().decode(stdout));
  } catch (_) {
    console.error(`Failed to parse manifest for ${entry.name}`);
  }
  rows.push(
    "",
    `## [${manifest.title}](https://github.com/pomdtr/sunbeam/tree/main/extensions/${entry.name})`,
    "",
    `${manifest.description}`
  );

  if (manifest.preferences?.length) {
    rows.push("", "### Preferences", "");

    for (const preference of manifest.preferences) {
      rows.push(`- \`${preference.name}\`: ${preference.title}`);
    }
  }

  rows.push("", "### Commands", "");

  for (const command of manifest.commands) {
    rows.push(`- \`${command.name}\`: ${command.title}`);
  }

  rows.push(
    "",
    "### Install",
    "",
    "```",
    `sunbeam extension install https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/${entry.name}`,
    "```"
  );
}

Deno.writeTextFileSync(
  path.join(dirname, "..", "www", "frontend", "catalog", "index.md"),
  rows.join("\n")
);
