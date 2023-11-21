import fs from "fs/promises"
import path from "path"
import { spawnSync } from "child_process"

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
const entries = await fs.readdir(extensionDir, { withFileTypes: true });
for (const entry of entries) {
    const entrypoint = path.join(extensionDir, entry.name);
    const { stdout, status } = spawnSync(entrypoint, {
        encoding: "utf-8",
    })
    if (status !== 0) {
        console.error(`Failed to run entrypoint for ${entry.name}`)
        process.exit(1)
    }

    let manifest
    try {
        manifest = JSON.parse(stdout);
    } catch (_) {
        console.error(`Failed to parse manifest for ${entry.name}`)
        process.exit(1)
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

await fs.writeFile(path.join(dirname, "catalog.md"), rows.join("\n"))
