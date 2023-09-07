import json2ts from "npm:json-schema-to-typescript";
import { build, emptyDir } from "$dnt/mod.ts";

for (const name of ["page", "manifest"]) {
  // compile from file
  const ts = await json2ts.compileFromFile(`../../schemas/${name}.schema.json`);
  Deno.writeTextFileSync(`./${name}.ts`, ts);
}

await emptyDir(`./npm`);
await build({
  entryPoints: [
    { name: ".", path: "./mod.ts" },
    { name: "./extension", path: "./extension.ts" },
    { name: "./manifest", path: "./manifest.ts" },
    { name: "./page", path: "./page.ts" },
  ],
  outDir: `./npm`,
  shims: {
    // see JS docs for overview and more options
    deno: true,
    custom: [
      {
        package: {
          name: "isomorphic-fetch",
          version: "3.0.0",
        },
        typesPackage: {
          name: "@types/isomorphic-fetch",
          version: "0.0.36",
        },
        globalNames: [
          {
            // for the `fetch` global...
            name: "fetch",
            // use the default export of node-fetch
            exportName: "default",
          },
        ],
      },
    ],
  },
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
