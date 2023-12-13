#!/usr/bin/env -S deno run -A

import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import { editor } from "https://deno.land/x/sunbeam/editor.ts";
import * as path from "https://deno.land/std/path/mod.ts";

const manifest = {
  title: "Gists",
  description: "Manage your gists",
  preferences: [
    {
      name: "token",
      title: "GitHub Personal Access Token",
      type: "text",
    },
  ],
  commands: [
    {
      name: "manage",
      title: "Search Gists",
      mode: "filter",
    },
    {
      name: "create",
      title: "Create Gist",
      mode: "tty",
      params: [
        {
          name: "filename",
          title: "Filename",
          placeholder: "gist.md",
          type: "text",
        },
        {
          name: "description",
          title: "Description",
          optional: true,
          placeholder: "Gist Description",
          type: "text",
        },
        {
          name: "public",
          title: "Public",
          type: "checkbox",
          label: "Whether the gist is public or not.",
        },
      ],
    },
    {
      name: "browse",
      title: "Browse Gist Files",
      hidden: true,
      mode: "filter",
      params: [{ name: "id", title: "Gist ID", type: "text" }],
    },
    {
      name: "view",
      title: "View Gist File",
      hidden: true,
      mode: "detail",
      params: [
        { name: "id", title: "Gist ID", type: "text" },
        { name: "filename", title: "Filename", type: "text" },
      ],
    },
    {
      name: "edit",
      title: "Edit Gist File",
      hidden: true,
      mode: "tty",
      params: [
        { name: "id", title: "Gist ID", type: "text" },
        { name: "filename", title: "Filename", type: "text" },
      ],
    },
    {
      name: "delete",
      hidden: true,
      title: "Delete Gist",
      mode: "silent",
      params: [{ name: "id", title: "Gist ID", type: "text" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
const githubToken = payload.preferences.token;
if (!githubToken) {
  console.error("No github token set");
  Deno.exit(1);
}

try {
  const res = await run(payload);
  if (res) {
    console.log(JSON.stringify(res));
  }
} catch (e) {
  console.error(e);
  Deno.exit(1);
}

async function run(payload: sunbeam.Payload<typeof manifest>) {
  switch (payload.command) {
    case "manage": {
      const resp = await fetchGithub("/gists");
      if (resp.status != 200) {
        throw new Error("Failed to fetch gists");
      }

      const gists = (await resp.json()) as any[];
      return {
        items: gists.map((gist) => ({
          title: Object.keys(gist.files)[0],
          subtitle: gist.description || "",
          accessories: [gist.public ? "Public" : "Private"],
          actions: [
            Object.keys(gist.files).length > 1
              ? {
                  type: "run",
                  title: "Browse Files",
                  command: "browse",
                  params: {
                    id: gist.id,
                  },
                }
              : {
                  type: "run",
                  title: "View File",
                  command: "view",
                  params: {
                    id: gist.id,
                    filename: Object.keys(gist.files)[0],
                  },
                },
            {
              type: "open",
              title: "Open in Browser",
              url: gist.html_url,
              exit: true,
            },
            {
              type: "copy",
              title: "Copy URL",
              key: "c",
              text: gist.html_url,
              exit: true,
            },
            {
              title: "Create Gist",
              key: "n",
              type: "run",
              command: "create",
            },
            {
              title: "Delete Gist",
              key: "d",
              type: "run",
              command: "delete",
              reload: true,
              params: {
                id: gist.id,
              },
            },
          ],
        })),
      } as sunbeam.List;
    }
    case "create": {
      const filename = payload.params.filename;
      const content = await editor({
        extension: path.extname(filename),
        content: "",
      });
      const resp = await fetchGithub("/gists", {
        method: "POST",
        body: JSON.stringify({
          description: payload.params.description,
          public: payload.params.public,
          files: {
            [filename]: {
              content,
            },
          },
        }),
      });

      if (resp.status != 201) {
        throw new Error("Failed to create gist");
      }
      return;
    }
    case "browse": {
      const id = payload.params.id;
      const resp = await fetchGithub(`/gists/${id}`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch gist");
      }

      const gist = (await resp.json()) as any;
      return {
        items: Object.entries(gist.files).map(([filename]) => ({
          title: filename,
          actions: [
            {
              title: "View",
              type: "run",
              command: "view",
              params: {
                id,
                filename,
              },
            },
            {
              title: "Edit",
              type: "run",
              command: "edit",
              params: {
                id,
                filename,
              },
            },
          ],
        })),
      } as sunbeam.List;
    }
    case "view": {
      const { id, filename } = payload.params;
      const resp = await fetchGithub(`/gists/${id}`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch gist");
      }

      const gist = (await resp.json()) as any;
      const file = gist.files[filename];
      if (!file) {
        throw new Error("File not found");
      }
      const lang = file.language?.toLowerCase();

      return {
        markdown:
          lang == "md"
            ? file.content
            : `\`\`\`${lang || ""}\n${file.content}\n\`\`\``,
        actions: [
          {
            title: "Edit File",
            type: "run",
            command: "edit",
            params: {
              id,
              filename,
            },
            reload: true,
          },
          {
            title: "Copy Content",
            key: "c",
            type: "copy",
            text: file.content,
            exit: true,
          },
          {
            title: "Copy Raw URL",
            key: "r",
            type: "copy",
            text: file.raw_url,
            exit: true,
          },
          {
            title: "Open in Browser",
            type: "open",
            url: gist.html_url,
            exit: true,
          },
        ],
      } as sunbeam.Detail;
    }
    case "edit": {
      const { id, filename } = payload.params;
      const get = await fetchGithub(`/gists/${id}`);
      if (get.status != 200) {
        throw new Error("Failed to fetch gist");
      }

      const gist = (await get.json()) as any;
      const file = gist.files[filename];
      if (!file) {
        throw new Error("File not found");
      }

      const extension = path.extname(filename);
      const content = await editor({ extension, content: file.content });
      if (content == file.content) {
        return;
      }

      const patch = await fetchGithub(`/gists/${id}`, {
        method: "PATCH",
        body: JSON.stringify({
          files: {
            [filename]: {
              content,
            },
          },
        }),
      });

      if (patch.status != 200) {
        throw new Error("Failed to update gist");
      }

      return;
    }
    case "delete": {
      const id = payload.params.id;
      const resp = await fetchGithub(`/gists/${id}`, {
        method: "DELETE",
      });
      if (resp.status != 204) {
        throw new Error("Failed to delete gist");
      }
    }
  }
}

function fetchGithub(endpoint: string, init?: RequestInit) {
  return fetch(`https://api.github.com${endpoint}`, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${githubToken}`,
    },
  });
}
