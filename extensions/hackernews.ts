#!/usr/bin/env -S deno run -A

import Parser from "npm:rss-parser";
import { formatDistance } from "npm:date-fns";
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Hacker News",
  description: "Browse Hacker News",
  root: [
    {
      title: "Hacker News Front Page",
      command: "browse",
      params: {
        topic: "frontpage",
      },
    },
  ],
  commands: [
    {
      name: "browse",
      title: "Show a feed",
      mode: "filter",
      params: [{ name: "topic", title: "Topic", type: "string" }],
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
    `https://hnrss.org/${topic}?description=0&count=25`,
  );

  const items: sunbeam.ListItem[] = feed.items?.map((item) => ({
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
        extension: "std",
        command: "open",
        params: {
          url: item.link || "",
        },
      },
      {
        title: "Open Comments in Browser",
        extension: "std",
        command: "open",
        params: {
          url: item.guid || "",
        },
      },
      {
        title: "Copy Link",
        extension: "std",
        command: "copy",
        params: {
          text: item.link || "",
        },
      },
    ],
  })) || [];

  const page: sunbeam.List = {
    items,
  };

  console.log(JSON.stringify(page));
} else {
  console.error("Unknown command");
  Deno.exit(1);
}
