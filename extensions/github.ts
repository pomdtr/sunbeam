#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as base64 from "https://deno.land/std@0.202.0/encoding/base64.ts";

const manifest = {
  title: "GitHub",
  description: "Search GitHub repositories",
  preferences: [
    {
      name: "token",
      description: "Personal Access Token",
      type: "string",
    },
  ],
  commands: [
    {
      title: "Search Repositories",
      name: "search",
      mode: "search",
    },
    {
      title: "List Issues",
      name: "issue.list",
      hidden: true,
      mode: "filter",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
    {
      title: "List Pull Requests",
      name: "pr.list",
      hidden: true,
      mode: "filter",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
    {
      title: "View Readme",
      name: "readme",
      hidden: true,
      mode: "detail",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest, null, 2));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
try {
  await run(payload);
} catch (err) {
  console.error(err);
  Deno.exit(1);
}

async function run(payload: sunbeam.Payload<typeof manifest>) {
  const token = payload.preferences.token;
  if (payload.command == "search") {
    const query = payload.query;
    if (!query) {
      const list: sunbeam.List = {
        emptyText: "Enter a query to search for repositories",
        items: [],
      };
      console.log(JSON.stringify(list, null, 2));
      return;
    }

    const resp = await fetch(
      `https://api.github.com/search/repositories?q=${
        encodeURIComponent(
          query,
        )
      }`,
      {
        headers: {
          Authorization: `token ${token}`,
        },
      },
    );

    if (!resp.ok) {
      throw new Error(
        `Failed to search repositories: ${resp.status} ${resp.statusText}`,
      );
    }

    const data = await resp.json();
    const list: sunbeam.List = {
      items: data.items.map(
        (item: any) => ({
          title: item.full_name,
          accessories: [`${item.stargazers_count} *`],
          actions: [
            {
              title: "View README",
              type: "run",
              command: "readme",
              params: {
                repo: item.full_name,
              },
            },
            {
              title: "Open In Browser",
              key: "o",
              type: "open",
              url: item.html_url,
              exit: true,
            },
            {
              title: "List Issues",
              key: "i",
              type: "run",
              command: "issue.list",
              params: {
                repo: item.full_name,
              },
            },
            {
              title: "List Pull Requests",
              key: "p",
              type: "run",
              command: "pr.list",
              params: {
                repo: item.full_name,
              },
            },
            {
              title: "Copy URL",
              key: "c",
              type: "copy",
              text: item.html_url,
              exit: true,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (payload.command == "issue.list") {
    const repo = payload.params.repo;
    const resp = await fetch(`https://api.github.com/repos/${repo}/issues`, {
      headers: {
        Authorization: `token ${token}`,
      },
    });

    if (!resp.ok) {
      throw new Error(
        `Failed to list issues: ${resp.status} ${resp.statusText}`,
      );
    }

    const data = await resp.json();
    const list: sunbeam.List = {
      items: data.map(
        (item: any) => ({
          title: item.title,
          accessories: [`#${item.number}`],
          actions: [
            {
              title: "Open In Browser",
              type: "open",
              url: item.html_url,
              exit: true,
            },
            {
              title: "Copy URL",
              key: "c",
              type: "copy",
              text: item.html_url,
              exit: true,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (payload.command == "pr.list") {
    const repo = payload.params.repo;
    const resp = await fetch(`https://api.github.com/repos/${repo}/pulls`, {
      headers: {
        Authorization: `token ${token}`,
      },
    });

    if (!resp.ok) {
      throw new Error(
        `Failed to list pull requests: ${resp.status} ${resp.statusText}`,
      );
    }

    const data = await resp.json();
    const list: sunbeam.List = {
      items: data.map(
        (item: any) => ({
          title: item.title,
          accessories: [`#${item.number}`],
          actions: [
            {
              title: "Open In Browser",
              type: "open",
              url: item.html_url,
              exit: true,
            },
            {
              title: "Copy URL",
              key: "c",
              type: "copy",
              text: item.html_url,
              exit: true,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (payload.command == "readme") {
    const repo = payload.params.repo;
    const resp = await fetch(`https://api.github.com/repos/${repo}/readme`, {
      headers: {
        Authorization: `token ${token}`,
      },
    });

    if (!resp.ok) {
      throw new Error(
        `Failed to view readme: ${resp.status} ${resp.statusText}`,
      );
    }

    const data = await resp.json();
    const markdown = new TextDecoder().decode(base64.decode(data.content));

    const detail: sunbeam.Detail = {
      markdown,
      actions: [
        {
          title: "Open in Browser",
          type: "open",
          url: data.html_url,
        },
        {
          title: "List Issues",
          key: "i",
          type: "run",
          command: "issue.list",
          params: {
            repo: repo,
          },
        },
        {
          title: "List Pull Requests",
          key: "p",
          type: "run",
          command: "pr.list",
          params: {
            repo: repo,
          },
        },
      ],
    };

    console.log(JSON.stringify(detail, null, 2));
  }
}
