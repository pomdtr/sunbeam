#!/usr/bin/env -S deno run -A
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
import * as clipboard from "https://deno.land/x/copy_paste@v1.1.3/mod.ts";

const manifest = {
  title: "Pipe Commands",
  description: "Pipe your clipboard through various commands",
  commands: [
    {
      name: "urldecode",
      title: "URL Decode Clipboard",
      mode: "silent",
    },
    {
      name: "urlencode",
      title: "URL Encode Clipboard",
      mode: "silent",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "urldecode") {
  const content = await clipboard.readText();
  const decoded = decodeURIComponent(content);
  await clipboard.writeText(decoded);
} else if (payload.command == "urlencode") {
  const content = await clipboard.readText();
  const encoded = encodeURIComponent(content);
  await clipboard.writeText(encoded);
}
