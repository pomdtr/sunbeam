#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "My Extension",
  description: "This is my extension",
  commands: [
    {
      name: "hi",
      title: "Say Hi",
      mode: "detail",
      params: [
        {
          name: "name",
          title: "Name",
          type: "text",
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
if (payload.command == "hi") {
  const name = payload.params.name;
  const detail: sunbeam.Detail = {
    text: `Hi ${name}!`,
    actions: [
      {
        title: "Copy Name",
        type: "copy",
        text: name,
      },
    ],
  };
  console.log(JSON.stringify(detail));
} else {
  console.error(`Unknown command: ${payload.command}`);
  Deno.exit(1);
}
