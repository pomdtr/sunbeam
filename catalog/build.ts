#!/usr/bin/env -S deno run -A
import { markdownTable } from 'npm:markdown-table';
import { join } from "https://deno.land/std/path/mod.ts";
import * as sunbeam from "npm:sunbeam-types@0.23.12";

const dirname = new URL(".", import.meta.url).pathname;

const entries = Deno.readDirSync(join(dirname, "extensions"));
const extensions: {
    entrypoint: string;
    title: string;
    description: string;
}[] = []
for (const entry of entries) {
    const entrypoint = join(dirname, "extensions", entry.name);
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
    ...extensions.map(({ entrypoint, title, description }) => [`[${title}](https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/${entrypoint})`, description])
]);

const template = await Deno.readTextFileSync(join(dirname, "README.tmpl.md"));
const readme = template.replace("{{catalog}}", table);
await Deno.writeTextFile(join(dirname, "README.md"), readme);
