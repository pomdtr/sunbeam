#!/usr/bin/env zx

const query = argv.query;
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

for (const item of items) {
  console.log(JSON.stringify(item));
}
