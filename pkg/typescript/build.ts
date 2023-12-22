#!/usr/bin/env -S deno run -A

import { build, emptyDir } from "https://deno.land/x/dnt/mod.ts";

await emptyDir("./npm");

const dirname = new URL(".", import.meta.url).pathname;
const version = Deno.readTextFileSync(`${dirname}/version.txt`).trim();

await build({
  entryPoints: ["./src/mod.ts"],
  outDir: "./npm",
  shims: {
    // see JS docs for overview and more options
    deno: true,
  },
  package: {
    // package.json properties
    name: "sunbeam-sdk",
    version,
    description: "Sunbeam Types and Utilities",
    license: "MIT",
    repository: {
      type: "git",
      url: "git+https://github.com/pomdtr/sunbeam.git",
    },
    bugs: {
      url: "https://github.com/pomdtr/sunbeam/issues",
    },
  },
  postBuild() {
    // steps to run after building and before running the tests
    Deno.copyFileSync(`${dirname}/README.md`, `${dirname}/npm/README.md`);
  },
});
