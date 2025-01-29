#!/usr/bin/env -S deno run -A

import * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.5";
import * as base64 from "https://deno.land/std@0.202.0/encoding/base64.ts";
import { toJson } from "jsr:@std/streams";

const manifest = {
  title: "GitHub",
  description: "Search GitHub repositories",
  commands: [
    {
      description: "Search Repositories",
      name: "search",
      mode: "search",
    },
    {
      description: "List Issues",
      name: "issue.list",
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
      description: "List Pull Requests",
      name: "pr.list",
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
      description: "View Readme",
      name: "readme",
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

const token = Deno.env.get("GITHUB_TOKEN");
if (!token) {
  console.error("No GitHub token set");
  Deno.exit(1);
}

try {
  await run(Deno.args[0], await toJson(Deno.stdin.readable) as sunbeam.Payload);
} catch (err) {
  console.error(err);
  Deno.exit(1);
}

async function run(command: string, payload: sunbeam.Payload) {
  if (command == "search") {
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
      `https://api.github.com/search/repositories?q=${encodeURIComponent(
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
              type: "open",
              target: item.html_url,
            },
            {
              title: "List Issues",
              type: "run",
              command: "issue.list",
              params: {
                repo: item.full_name,
              },
            },
            {
              title: "List Pull Requests",
              type: "run",
              command: "pr.list",
              params: {
                repo: item.full_name,
              },
            },
            {
              title: "Copy URL",
              type: "copy",
              text: item.html_url,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (command == "issue.list") {
    const repo = payload.repo;
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
              target: item.html_url,
            },
            {
              title: "Copy URL",
              type: "copy",
              text: item.html_url,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (command == "pr.list") {
    const repo = payload.repo;
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
              target: item.html_url,
            },
            {
              title: "Copy URL",
              key: "c",
              type: "copy",
              text: item.html_url,
            },
          ],
        } as sunbeam.ListItem),
      ),
    };

    console.log(JSON.stringify(list, null, 2));
  } else if (command == "readme") {
    const repo = payload.repo;
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
          target: data.html_url,
        },
        {
          title: "List Issues",
          type: "run",
          command: "issue.list",
          params: {
            repo: repo,
          },
        },
        {
          title: "List Pull Requests",
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
