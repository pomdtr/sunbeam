#!/usr/bin/env -S deno run -A
import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Raindrop",
  description: "Manage your raindrop bookmarks",
  preferences: [
    {
      name: "token",
      title: "Raindrop API Token",
      type: "text",
    },
  ],
  root: ["search-bookmarks"],
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
const raindropToken = payload.preferences.token;
if (!raindropToken) {
  console.error("No raindrop token found, please set it in your config");
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
          type: "open",
          url: bookmark.link,
          exit: true,
        },
        {
          title: "Copy URL",
          key: "c",
          type: "copy",
          text: bookmark.link,
          exit: true,
        },
      ],
    })),
  };

  console.log(JSON.stringify(list));
} else {
  console.error(`Unknown command: ${payload.command}`);
  Deno.exit(1);
}
