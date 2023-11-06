#!/usr/bin/env deno run -A
import * as sunbeam from "npm:sunbeam-types@0.23.19"
import * as clipboard from "https://deno.land/x/copy_paste@v1.1.3/mod.ts";


if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Pipe Commands",
        description: "Pipe your clipboard through various commands",
        requirements: [
            {
                name: "deno",
                link: "https://deno.com"
            }
        ],
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
            }
        ]
    }

    console.log(JSON.stringify(manifest))
    Deno.exit(0)
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload
if (payload.command == "urldecode") {
    const content = await clipboard.readText()
    const decoded = decodeURIComponent(content)
    await clipboard.writeText(decoded)
} else if (payload.command == "urlencode") {
    const content = await clipboard.readText()
    const encoded = encodeURIComponent(content)
    await clipboard.writeText(encoded)
}
