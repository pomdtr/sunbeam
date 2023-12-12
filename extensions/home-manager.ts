#!/usr/bin/env -S deno run -A

import { parse } from "npm:node-html-parser";
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest: sunbeam.Manifest = {
  title: "Home Manager",
  root: ["search"],
  commands: [
    {
      name: "search",
      title: "Search Configuration Options",
      mode: "filter",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const baseUrl = "https://nix-community.github.io/home-manager/options.html";
const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "search") {
  const resp = await fetch(baseUrl);
  const html = await resp.text();
  const res = parse(html);
  const dts = res.querySelectorAll("dl dt");
  const dds = res.querySelectorAll("dl dd");
  const items: sunbeam.ListItem[] = dts.map((el, idx) => ({
    title: el.querySelector("code")?.text || "",
    subtitle: dds[idx].childNodes[0].text,
    actions: [
      {
        title: "Open in Browser",
        type: "open",
        url: new URL(
          el.querySelector("a.term")?.attributes["href"] || "",
          baseUrl
        ).toString(),
      },
      {
        title: "Copy Full Path",
        key: "c",
        exit: true,
        type: "copy",
        text: el.querySelector("code")?.text || "",
      },
      {
        title: "Copy Name",
        key: "n",
        exit: true,
        type: "copy",
        text: el.querySelector("code")?.text.split(".").slice(-1)[0] || "",
      },
    ],
  }));

  const list: sunbeam.List = {
    items,
  };

  console.log(JSON.stringify(list));
} else {
  throw new Error("Unknown command");
}
