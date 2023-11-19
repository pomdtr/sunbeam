#!/usr/bin/env deno run -A

import * as path from "https://deno.land/std@0.205.0/path/mod.ts";
import * as sunbeam from "../pkg/typescript/src/manifest.ts";

const dirname = new URL(".", import.meta.url).pathname;
const rows = []

rows.push(
    "---",
    "sidebar: false",
    "outline: 2",
    "---",
    "",
    "# Extension Catalog"
)

const extensionDir = path.join(dirname, "..", "extensions");
const entries = Deno.readDirSync(extensionDir);
for (const entry of entries) {
    const entrypoint = path.join(extensionDir, entry.name);
    const command = new Deno.Command(entrypoint)
    const { stdout } = await command.output()

    let manifest: sunbeam.Manifest
    try {
        manifest = JSON.parse(new TextDecoder().decode(stdout));
    } catch (_) {
        console.error(`Failed to parse manifest for ${entry.name}`)
        Deno.exit(1)
    }
    rows.push(
        "",
        `## [${manifest.title}](https://github.com/pomdtr/sunbeam/tree/main/extensions/${entry.name})`,
        "",
        `${manifest.description}`,
    )

    if (manifest.preferences?.length) {
        rows.push(
            "",
            "### Preferences",
            ""
        )

        for (const preference of manifest.preferences) {
            rows.push(
                `- \`${preference.name}\`: ${preference.title}`
            )
        }
    }

    rows.push(
        "",
        "### Commands",
        ""
    )

    for (const command of manifest.commands) {
        if (command.hidden) continue
        rows.push(
            `- \`${command.name}\`: ${command.title}`
        )
    }

    rows.push(
        "",
        "### Install",
        "",
        "```",
        `sunbeam extension install https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/${entry.name}`,
        "```"
    )
}

console.log(rows.join("\n"))
