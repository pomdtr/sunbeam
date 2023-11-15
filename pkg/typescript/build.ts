#!/usr/bin/env -S deno run -A
import { build, emptyDir } from "https://deno.land/x/dnt@0.39.0/mod.ts";

await emptyDir("./npm");
await build({
    entryPoints: ["./src/mod.ts"],
    outDir: "./npm",
    shims: {},
    package: {
        // package.json properties
        name: "sunbeam-types",
        version: Deno.readTextFileSync("./version.txt").trimEnd(),
        description: "Sunbeam Types",
        license: "MIT",
        repository: {
            type: "git",
            url: "git+https://github.com/pomdtr/sunbeam.git",
        },
        bugs: {
            url: "https://github.com/pomdtr/sunbeam/issues",
        },
    }
});
