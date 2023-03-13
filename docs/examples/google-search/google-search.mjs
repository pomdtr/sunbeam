#!/usr/bin/env zx

const query = await stdin();

let items = [];
if (query) {
  const res = await fetch(
    `https://www.google.com/complete/search?client=chrome&q=${encodeURIComponent(
      query
    )}}`
  ).then((res) => res.json());

  items = res[1].map((suggestion) => ({
    title: suggestion,
    actions: [
      {
        type: "open",
        url: `https://www.google.com/search?q=${encodeURIComponent(
          suggestion
        )}`,
      },
      {
        type: "copy",
        title: "Copy Suggestion",
        text: suggestion,
      },
    ],
  }));
}

console.log(
  JSON.stringify({
    type: "list",
    emptyText: query ? "No results" : "Enter a query",
    generateItems: true,
    items,
  })
);
