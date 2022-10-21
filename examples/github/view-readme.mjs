#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title View Readme
// @sunbeam.subtitle Github
// @sunbeam.mode interactive

// @sunbeam.argument1 { "type": "text", "placeholder": "repository", "required": true }

const repo = argv._[0];

const res = await fetch(
  `https://raw.githubusercontent.com/${repo}/master/README.md`
);
const readme = await res.text();
console.log(readme);
