#!/usr/bin/env -S deno run -A

import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as dates from "npm:date-fns";

const manifest = {
  title: "Deno Deploy",
  description: "Manage your Deno Deploy projects",
  preferences: [
    {
      name: "token",
      title: "Access Token",
      type: "text",
    },
  ],
  commands: [
    {
      name: "projects",
      title: "List Projects",
      mode: "filter",
    },
    {
      name: "dashboard",
      title: "Open Dashboard",
      mode: "silent",
    },
    {
      name: "deployments",
      title: "List Deployments",
      mode: "filter",
      params: [{ name: "project", title: "Project", type: "text" }],
    },
    {
      name: "playground",
      title: "View Playground",
      mode: "detail",
      params: [{ name: "project", title: "Project", type: "text" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
const deployToken = payload.preferences.token;
if (!deployToken) {
  console.error("Missing deploy token");
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
    case "dashboard": {
      await new Deno.Command("sunbeam", {
        args: ["open", "https://dash.deno.com"],
      }).output();
      return;
    }
    case "projects": {
      const resp = await fetchDeployAPI("/projects");
      if (resp.status != 200) {
        throw new Error("Failed to fetch projects");
      }
      const projects = await resp.json();

      return {
        items: projects.map((project: any) => {
          const item: sunbeam.ListItem = {
            title: project.name,
            accessories: [project.type],
            actions: [
              {
                title: "Open Dashboard",
                type: "open",
                url: `https://dash.deno.com/projects/${project.id}`,
              },
            ],
          };

          if (project.hasProductionDeployment) {
            const domains =
              project.productionDeployment.deployment.domainMappings;
            const domain = domains.length
              ? domains[domains.length - 1].domain
              : "No domain";
            item.subtitle = domain;
            item.actions?.push({
              title: "Open Production URL",
              type: "open",
              url: `https://${domain}`,
            });
          }

          item.actions?.push({
            title: "Copy Dashboard URL",
            type: "copy",
            key: "c",
            text: `https://dash.deno.com/projects/${project.id}`,
            exit: true,
          });

          return item;
        }),
      } as sunbeam.List;
    }
    case "playground": {
      const resp = await fetchDeployAPI(`/projects/${payload.params.project}`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch project");
      }

      const project = await resp.json();
      if (project.type != "playground") {
        throw new Error("Project is not a playground");
      }

      const snippet = project.playground.snippet;
      const lang = project.playground.mediaType;
      return {
        markdown: `\`\`\`${lang}\n${snippet}\n\`\`\``,
        actions: [
          {
            title: "Copy Snippet",
            key: "c",
            type: "copy",
            text: snippet,
            exit: true,
          },
          {
            title: "Open in Browser",
            key: "o",
            type: "open",
            url: `https://dash.deno.com/playground/${project.id}`,
            exit: true,
          },
        ],
      } as sunbeam.Detail;
    }
    case "deployments": {
      const project = payload.params.project;

      const resp = await fetchDeployAPI(`/projects/${project}/deployments`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch deployments");
      }

      const [deployments] = await resp.json();
      return {
        items: deployments.map(
          ({ id, createdAt, deployment, relatedCommit }: any) => {
            const item = {
              title: id,
              accessories: [
                dates.formatDistance(new Date(createdAt), new Date(), {
                  addSuffix: true,
                }),
              ],
              actions: [],
            } as sunbeam.ListItem;

            if (deployment.domainMappings?.length) {
              item.actions?.push({
                title: "Open URL",
                type: "open",
                url: `https://${deployment.domainMappings[0].domain}`,
              });
            }

            if (relatedCommit) {
              item.title = relatedCommit.message;
              item.actions?.push({
                title: "Open Commit",
                type: "open",
                url: relatedCommit.url,
              });
            }

            return item;
          }
        ),
      } as sunbeam.List;
    }
  }
}

function fetchDeployAPI(endpoint: string, init?: RequestInit) {
  return fetch(`https://dash.deno.com/api${endpoint}`, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${deployToken}`,
    },
  });
}
