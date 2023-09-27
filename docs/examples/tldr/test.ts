#!/usr/bin/env -S deno run -A

if (Deno.args.length == 0) {
  console.log(JSON.stringify({
    title: "tldr",
    commands: [
      {
        name: "tldr",
        title: "List Tldr pages",
        mode: "view",
      },
    ],
  }));
  Deno.exit(0);
}

const { command, params } = JSON.parse(Deno.args[0]);
switch (command) {
  case "tldr":
    console.log(JSON.stringify({
      type: "list",
      items: [
        { title: "Item 1", subtitle: JSON.stringify(params) },
      ],
    }));
}
