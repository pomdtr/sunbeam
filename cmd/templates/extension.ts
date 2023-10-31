#!/usr/bin/env -S deno run -A

import * as sunbeam from "npm:sunbeam-types@0.23.15"

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Example Extension",
        description: "Example extension",
        commands: [
            {
                name: "hello",
                title: "Hello",
                mode: "detail"
            }
        ]
    }

    console.log(JSON.stringify(manifest))
    Deno.exit()
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.CommandInput

if (payload.command == "hello") {
    const detail: sunbeam.Detail = {
        text: "Hello, world!"
    }

    console.log(JSON.stringify(detail))
}
