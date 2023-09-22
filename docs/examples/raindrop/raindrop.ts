#!/usr/bin/env -S deno run -A
// deno-lint-ignore-file no-explicit-any

import * as sunbeam from "../../../pkg/typescript/mod.ts";
import exec from "../../../pkg/typescript/exec.ts";

const token = Deno.env.get("RAINDROP_TOKEN");
if (!token) {
  throw new Error("RAINDROP_TOKEN is required");
}

function fetchRaindrop(path: string, init?: RequestInit) {
  const url = new URL(`https://api.raindrop.io/rest/v1/${path}`);
  return fetch(url, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${token}`,
    },
  });
}

const extension = new sunbeam.ExtensionClass({
  title: "Raindrop",
  description: "Search your Raindrop bookmarks",
}).command({
  name: "edit-raindrop-form",
  title: "Edit a bookmark",
  params: {
    "bookmark": {
      name: "bookmark",
      type: "string",
    },
  },
  output: "form",
  run: async ({ params }) => {
    const resp = await fetchRaindrop(`raindrops/${params.bookmark}`);
    if (!resp.ok) {
      throw new Error(`Failed to fetch bookmark: ${resp.statusText}`);
    }

    const { item } = await resp.json();
    return {
      command: {
        name: "edit-raindrop",
        params: {
          bookmark: params.bookmark,
        },
      },
      items: [
        {
          type: "textfield",
          name: "title",
          title: "Title",
          default: item.title,
        },
        { type: "textfield", name: "link", title: "URL", value: item.link },
      ],
    } as sunbeam.Form;
  },
}).command({
  name: "edit-raindrop",
  title: "Edit a bookmark",
  params: {
    "bookmark": {
      type: "string",
    },
    "title": {
      type: "string",
    },
    "link": {
      type: "string",
    },
  },
  run: async () => {},
}).command({
  name: "search",
  title: "Search Raindrop Bookmarks",
  params: {
    collection: {
      type: "string",
      "optional": true,
    },
  },
  output: "list",
  run: async ({ params }) => {
    const collectionId = params.collection as string || "0";
    const resp = await fetchRaindrop(`raindrops/${collectionId}`);
    if (!resp.ok) {
      throw new Error(`Failed to fetch bookmarks: ${resp.statusText}`);
    }
    const { items: bookmarks } = await resp.json();

    return {
      items: bookmarks.map((bookmark: any) => ({
        title: bookmark.title,
        subtitle: bookmark.link || "",
        actions: [
          {
            title: "Open in Raindrop",
            type: "open",
            url: `https://app.raindrop.io/my/0/item/${bookmark._id}/preview`,
          },
          {
            title: "Open in Browser",
            type: "open",
            url: bookmark.link,
          },
          {
            title: "Copy URL",
            type: "copy",
            text: bookmark.link,
          },
          {
            title: "Edit",
            type: "run",
            command: {
              name: "edit-raindrop-form",
              params: {
                bookmark: bookmark._id,
              },
            },
          },
          {
            title: "Delete",
            type: "run",
            command: {
              name: "delete-raindrop",
              params: {
                bookmark: bookmark._id,
              },
            },
          },
          {
            title: "Add Bookmark",
            type: "run",
            command: {
              name: "add-raindrop-form",
            },
          },
        ],
      })),
    };
  },
}).command({
  name: "add-raindrop-form",
  title: "Add a bookmark",
  output: "form",
  run: () => {
    return {
      command: {
        name: "create-raindrop",
      },
      items: [
        { type: "textfield", name: "title", title: "Title" },
        { type: "textfield", name: "link", title: "URL" },
      ],
    } as sunbeam.Form;
  },
}).command({
  name: "create-raindrop",
  title: "Add a bookmark",
  params: {
    title: {
      type: "string",
    },
    link: {
      type: "string",
    },
  },
  run: async ({ params }) => {
    const resp = await fetchRaindrop("raindrops", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        title: params.title,
        link: params.link,
      }),
    });
    if (!resp.ok) {
      throw new Error(`Failed to create bookmark: ${resp.statusText}`);
    }
  },
}).command({
  name: "delete-raindrop",
  title: "Delete a bookmark",
  params: {
    bookmark: {
      type: "string",
    },
  },
  run: async ({ params }) => {
    const resp = await fetchRaindrop(`raindrops/${params.bookmark}`, {
      method: "DELETE",
    });
    if (!resp.ok) {
      throw new Error(`Failed to delete bookmark: ${resp.statusText}`);
    }
  },
});

exec(extension);
