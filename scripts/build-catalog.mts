#!/usr/bin/env deno run -A

import { markdownTable } from 'npm:markdown-table';
import * as path from "https://deno.land/std@0.205.0/path/mod.ts";
import * as sunbeam from "npm:sunbeam-types@0.23.16";

const dirname = new URL(".", import.meta.url).pathname;

console.log(`---
sidebar: false
outline: 2
---

# Extension Catalog
`)

const extensionDir = path.join(dirname, "..", "extensions");
const entries = Deno.readDirSync(extensionDir);
for (const entry of entries) {
    const entrypoint = path.join(extensionDir, entry.name);
    const command = new Deno.Command(entrypoint)
    const { stdout } = await command.output()

    const manifest: sunbeam.Manifest = JSON.parse(new TextDecoder().decode(stdout));

    console.log(`## [${manifest.title}](https://github.com/pomdtr/sunbeam/tree/main/extensions/${entry.name})

${manifest.description}

### Commands

${manifest.commands.map((command) => (
        `- \`${command.name}\`: ${command.title}`
    )).join("\n")}

### Installation

\`\`\`
sunbeam extension install https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/${entry.name}
\`\`\`
`)
}
