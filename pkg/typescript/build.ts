import json2ts from "npm:json-schema-to-typescript";
import { build, emptyDir } from "$dnt/mod.ts";

for (const name of ["page", "manifest"]) {
  // compile from file
  const ts = await json2ts.compileFromFile(`../schemas/${name}.schema.json`);
  Deno.writeTextFileSync(`./${name}.ts`, ts);
}

await emptyDir(`./npm`);
await build({
  entryPoints: ["./mod.ts"],
  outDir: `./npm`,
  shims: {},
  package: {
    // package.json properties
    name: "sunbeam-types",
    version: Deno.args[0],
    description: "Types for sunbeam.",
    license: "MIT",
    repository: {
      type: "git",
      url: "git+https://github.com/pomdtr/sunbeam.git",
    },
    bugs: {
      url: "https://github.com/pomdtr/sunbeam/issues",
    },
  },
});
