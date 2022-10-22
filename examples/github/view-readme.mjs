#!/usr/bin/env zx

const repo = argv._[0];

const res = await fetch(
  `https://raw.githubusercontent.com/${repo}/master/README.md`
);
const readme = await res.text();
console.log(readme);
