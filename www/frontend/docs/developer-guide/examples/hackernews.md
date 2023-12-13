---
title: Hacker News
---

```ts
#!/usr/bin/env -S deno run -A

import Parser from "npm:rss-parser";
import { formatDistance } from "npm:date-fns";
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Hacker News",
  description: "Browse Hacker News",
  commands: [
    {
      name: "browse",
      title: "Show a feed",
      mode: "filter",
      params: [{ name: "topic", title: "Topic", type: "text" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "browse") {
  const { topic } = payload.params;
  const feed = await new Parser().parseURL(
    `https://hnrss.org/${topic}?description=0&count=25`
  );
  const page: sunbeam.List = {
    items: feed.items.map((item) => ({
      title: item.title || "",
      subtitle: item.categories?.join(", ") || "",
      accessories: item.isoDate
        ? [
            formatDistance(new Date(item.isoDate), new Date(), {
              addSuffix: true,
            }),
          ]
        : [],
      actions: [
        {
          title: "Open in browser",
          type: "open",
          url: item.link || "",
          exit: true,
        },
        {
          title: "Open Comments in Browser",
          type: "open",
          url: item.guid || "",
          exit: true,
        },
        {
          title: "Copy Link",
          type: "copy",
          key: "c",
          text: item.link || "",
          exit: true,
        },
      ],
    })),
  };

  console.log(JSON.stringify(page));
} else {
  console.error("Unknown command");
  Deno.exit(1);
}
```

You can add new sections by adding new items in your config.

```json
{
    "extensions": {
        "hackernews": {
            "origin": "...",
            "root": [ {
                "title": "Show HN",
                "command": "browse",
                "params": {
                    "topic": "show"
                }
            }]
        }
    }
}
```
