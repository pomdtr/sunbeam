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
      type: "openUrl",
      shortcut: "enter",
      url: `https://www.google.com/search?q=${encodeURIComponent(suggestion)}`,
    },
    {
      type: "copyText",
      title: "Copy Suggestion",
      shortcut: "ctrl+y",
      text: suggestion,
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
