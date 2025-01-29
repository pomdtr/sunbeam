#!/usr/bin/env -S deno run -A

import type * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.5";
import * as dates from "npm:date-fns";
import { toJson } from "jsr:@std/streams";

const manifest = {
  title: "Deno Deploy",
  description: "Manage your Deno Deploy projects",
  commands: [
    {
      name: "projects",
      description: "List Projects",
      mode: "filter",
    },
    {
      name: "dashboard",
      description: "Open Dashboard",
      mode: "silent",
    },
    {
      name: "deployments",
      description: "List Deployments",
      mode: "filter",
      params: [{ name: "project", type: "string" }],
    },
    {
      name: "playground",
      description: "View Playground",
      mode: "detail",
      params: [{ name: "project", type: "string" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const { DENO_DEPLOY_TOKEN } = Deno.env.toObject();
try {
  const command = Deno.args[0];
  const params = await toJson(Deno.stdin.readable) as sunbeam.Payload

  const res = await run(command, params);
  if (res) {
    console.log(JSON.stringify(res));
  }
} catch (e) {
  console.error(e);
  Deno.exit(1);
}

async function run(command: string, params: Record<string, any>) {
  switch (command) {
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
                target: `https://dash.deno.com/projects/${project.id}`,
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
              target: `https://${domain}`,
            });
          }

          item.actions?.push({
            title: "Copy Dashboard URL",
            type: "copy",
            text: `https://dash.deno.com/projects/${project.id}`,
          });

          return item;
        }),
      } as sunbeam.List;
    }
    case "playground": {
      const resp = await fetchDeployAPI(`/projects/${params.project}`);
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
            type: "copy",
            text: snippet,
          },
          {
            title: "Open in Browser",
            type: "open",
            target: `https://dash.deno.com/playground/${project.id}`,
          },
        ],
      } as sunbeam.Detail;
    }
    case "deployments": {
      const project = params.project;

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
                target: `https://${deployment.domainMappings[0].domain}`,
              });
            }

            if (relatedCommit) {
              item.title = relatedCommit.message;
              item.actions?.push({
                title: "Open Commit",
                type: "open",
                target: relatedCommit.url,
              });
            }

            return item;
          },
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
      Authorization: `Bearer ${DENO_DEPLOY_TOKEN}`,
    },
  });
}
