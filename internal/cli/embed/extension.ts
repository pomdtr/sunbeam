#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "My Extension",
        description: "This is my extension",
        items: [
            {
                title: "Hi Mom!",
                command: "hi",
                params: {
                    name: "Mom"
                }
            },
            {
                command: "hi",
            }
        ],
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
                    }
                ]
            }
        ]
    }
    console.log(JSON.stringify(manifest));
    Deno.exit(0)
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;

if (payload.command == "hi") {
    const name = payload.params.name as string;
    const detail: sunbeam.Detail = {
        text: `Hi ${name}!`,
        actions: [
            {
                title: "Copy Name",
                type: "copy",
                text: name
            }
        ]
    }
    console.log(JSON.stringify(detail));
} else {
    console.error(`Unknown command: ${payload.command}`);
    Deno.exit(1);
}


