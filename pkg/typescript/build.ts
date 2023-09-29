import json2ts from "npm:json-schema-to-typescript";
import { build, emptyDir } from "https://deno.land/x/dnt@0.38.0/mod.ts";

// compile from file
for (const name of ["manifest", "page", "config"]) {
  const ts = await json2ts.compileFromFile(`../schemas/${name}.schema.json`, {
    cwd: "../schemas",
  });
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
