#!/usr/bin/env zx

const query = argv.query;
if (!query) {
  console.log(
    JSON.stringify({
      title: "Please enter a query",
    })
  );
  process.exit();
}

const res = await fetch(
  `https://www.google.com/complete/search?client=chrome&q=${encodeURIComponent(
    query
  )}`
).then((res) => res.json());

const items = res[1].map((suggestion) => ({
  title: suggestion,
  actions: [
    {
      type: "open",
      url: `https://www.google.com/search?q=${encodeURIComponent(suggestion)}`,
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
