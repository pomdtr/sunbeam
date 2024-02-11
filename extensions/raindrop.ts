#!/usr/bin/env -S deno run -A
import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Raindrop",
  description: "Manage your raindrop bookmarks",
  commands: [
    {
      title: "Search Bookmarks",
      name: "search-bookmarks",
      mode: "filter",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length === 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
const raindropToken = Deno.env.get("RAINDROP_TOKEN");
if (!raindropToken) {
  console.error("RAINDROP_TOKEN env var is required.");
  Deno.exit(1);
}

if (payload.command == "search-bookmarks") {
  const resp = await fetch("https://api.raindrop.io/rest/v1/raindrops/0", {
    headers: {
      Authorization: `Bearer ${raindropToken}`,
    },
  });

  const { items: bookmarks } = (await resp.json()) as {
    items: {
      title: string;
      link: string;
      domain: string;
    }[];
  };

  const list: sunbeam.List = {
    items: bookmarks.map((bookmark) => ({
      title: bookmark.title,
      subtitle: bookmark.domain,
      actions: [
        {
          title: "Open URL",
          extension: "std",
          command: "open",
          params: {
            url: bookmark.link,
          },
        },
        {
          title: "Copy URL",
          key: "c",
          extension: "std",
          command: "copy",
          params: {
            text: bookmark.link,
          },
        },
      ],
    })),
  };

  console.log(JSON.stringify(list));
} else {
  console.error(`Unknown command: ${payload.command}`);
  Deno.exit(1);
}
