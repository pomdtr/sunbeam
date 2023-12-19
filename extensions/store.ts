#!/usr/bin/env -S deno run -A

import type * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import $ from "https://deno.land/x/dax/mod.ts";

const manifest = {
  title: "Store",
  description: "Download and install extensions",
  commands: [
    {
      title: "Search",
      name: "search",
      mode: "filter",
      params: [
        {
          name: "repository",
          label: "Repository",
          type: "text",
          text: {
            placeholder: "pomdtr/sunbeam",
          },
        },
      ],
    },
    {
      title: "Install",
      name: "install",
      mode: "detail",
      params: [
        {
          name: "origin",
          label: "Origin URL",
          type: "text",
        },
        {
          name: "alias",
          label: "Extension Alias",
          type: "text",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length === 0) {
  console.log(JSON.stringify(manifest, null, 2));
  Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload<typeof manifest>;

if (payload.command === "search") {
  const { repository } = payload.params;
  const rawRoot = `https://raw.githubusercontent.com/${repository}/main/`;
  const { extensions }: sunbeam.Config = await fetch(
    `${rawRoot}sunbeam.json`
  ).then((res) => res.json());

  const list: sunbeam.List = {
    items: Object.entries(extensions || {}).map(([alias, extension]) => {
      const origin = new URL(extension.origin, rawRoot).toString();
      return {
        title: alias,
        accessories: [origin.toString()],
        actions: [
          {
            type: "run",
            title: "Install",
            run: {
              command: "install",
              params: {
                origin: origin,
                alias: {
                  default: alias,
                },
              },
            },
          },
          {
            type: "copy",
            title: "Copy Install Command",
            copy: {
              exit: true,
              text: `sunbeam extension install --alias ${alias} '${origin}'`,
            },
          },
          {
            type: "open",
            title: "Open in Browser",
            open: {
              url: new URL(
                extension.origin,
                `https://github.com/${repository}/tree/main/`
              ).toString(),
            },
          },
        ],
      };
    }),
  };

  console.log(JSON.stringify(list));
} else if (payload.command === "install") {
  const { origin, alias } = payload.params;
  await $`sunbeam extension install --alias ${alias} ${origin}`;

  const detail: sunbeam.Detail = {
    markdown: `Installed \`${alias}\` from \`${origin}\``,
    actions: [
      {
        type: "exit",
        title: "Close",
      },
    ],
  };
  console.log(JSON.stringify(detail));
}
