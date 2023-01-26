#!/usr/bin/env zx

const query = await stdin();
if (!query) {
  console.log(
    JSON.stringify({
      title: "Please enter a query",
    })
  );
  process.exit();
}

const res = await fetch(
  `https://www.google.com/complete/search?client=chrome&q=${query}}`
).then((res) => res.json());

const items = res[1].map((suggestion) => ({
  title: suggestion,
  actions: [
    {
      type: "open-url",
      url: `https://www.google.com/search?q=${encodeURIComponent(suggestion)}`,
    },
    {
      type: "copy-text",
      title: "Copy Suggestion",
      shortcut: "ctrl+y",
      text: suggestion,
    },
  ],
}));

console.log(
  JSON.stringify({
    type: "list",
    isGenerator: true,
    items,
  })
);
