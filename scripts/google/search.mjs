#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Google Search
// @sunbeam.packageName Google Search

const { query } = JSON.parse(await stdin());

const payload = await fetch(
  `https://www.google.com/complete/search?client=chrome&q=${query}`
).then((res) => res.json());

const items = payload[1].map((suggestion) => ({
  title: suggestion,
  actions: [
    {
      type: "open-url",
      url: `https://www.google.com/search?q=${suggestion}`,
    },
  ],
}));

console.log(
  JSON.stringify({
    type: "list",
    list: { onQueryChange: { type: "reload" }, items },
  })
);
