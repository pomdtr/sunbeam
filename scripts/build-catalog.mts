#!/usr/bin/env deno run -A

import { markdownTable } from 'npm:markdown-table';
import * as path from "https://deno.land/std@0.205.0/path/mod.ts";
import * as sunbeam from "npm:sunbeam-types@0.23.16";

const dirname = new URL(".", import.meta.url).pathname;

const extensionDir = path.join(dirname, "..", "extensions");
const entries = Deno.readDirSync(extensionDir);
const extensions: {
    entrypoint: string;
    title: string;
    description: string;
}[] = []
for (const entry of entries) {
    const entrypoint = path.join(extensionDir, entry.name);
    const command = new Deno.Command(entrypoint)
    console.error(`Loading manifest from ${entry.name}`)
    const { stdout } = await command.output()

    const manifest: sunbeam.Manifest = JSON.parse(new TextDecoder().decode(stdout));
    extensions.push({
        entrypoint,
        title: manifest.title,
        description: manifest.description || "",
    })
}

const table = markdownTable([
    ["Extension", "Description"],
    ...extensions.map(({ entrypoint, title, description }) => [`[${title}](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/extensions/${path.basename(entrypoint)})`, description])
]);

const template = await Deno.readTextFileSync(path.join(dirname, "catalog.tmpl.md"));
const readme = template.replace("{{catalog}}", table);
await Deno.writeTextFile(path.join(dirname, "..", "docs", "catalog.md"), readme);
