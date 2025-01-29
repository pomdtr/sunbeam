#!/usr/bin/env -S deno run -A

import Parser from "npm:rss-parser@3.9.0";
import { formatDistance } from "npm:date-fns@2.30.0";
import * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.2";

const manifest = {
  title: "RSS",
  description: "Manage your RSS feeds",
  commands: [
    {
      name: "show",
      description: "Show a Feed",
      mode: "filter",
      params: [
        {
          name: "url",
          title: "URL",
          type: "string",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "show") {
  const feed = await new Parser().parseURL(payload.params.url);
  const page: sunbeam.List = {
    items: feed.items?.map((item) => ({
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
          target: item.link || "",
        },
        {
          title: "Copy Link",
          type: "copy",
          key: "c",
          text: item.link || "",
        },
      ],
    })),
  };

  console.log(JSON.stringify(page));
}
