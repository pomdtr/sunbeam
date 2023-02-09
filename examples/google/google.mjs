#!/usr/bin/env zx

const query = argv._[0];

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
        type: "open-url",
        url: `https://www.google.com/search?q=${encodeURIComponent(
          suggestion
        )}`,
      },
      {
        type: "copy-text",
        title: "Copy Suggestion",
        shortcut: "ctrl+y",
        text: suggestion,
      },
    ],
  }));
}

console.log(
  JSON.stringify({
    type: "list",
    generateItems: true,
    emptyMessage: query ? "No results" : "Enter a query",
    items,
  })
);
