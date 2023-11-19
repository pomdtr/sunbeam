#!/usr/bin/env deno run --allow-net --allow-read --allow-env

import { HTTPException, Hono } from "https://deno.land/x/hono@v3.10.1/mod.ts"
import { serveStatic } from "https://deno.land/x/hono@v3.10.1/middleware.ts"
import { installScript } from "./install.ts"

const githubToken = Deno.env.get("GITHUB_TOKEN")
if (!githubToken) {
    throw new Error("GITHUB_TOKEN environment variable is required")
}

const app = new Hono()

app.get('/install.sh', async (c) => {
    let tag = c.req.query("tag");
    if (!tag) {
        const resp = await fetch("https://api.github.com/repos/pomdtr/sunbeam/releases", {
            headers: {
                "User-Agent": "install-sunbeam.deno.dev",
                "Accept": "application/vnd.github.v3+json",
                "Authorization": `Bearer ${githubToken}`
            }
        })
        const releases = await resp.json()
        if (releases.length == 0) {
            throw new HTTPException(404, { message: "No releases found" })
        }
        tag = releases[0].tag_name as string
    }

    return c.text(installScript(tag), {
        headers: {
            "Content-Disposition": `attachment; filename="install-sunbeam-${tag}.sh"`,
            "Content-Type": "application/x-shellscript"
        }
    })
})

app.get('*', serveStatic({ root: './static' }))

Deno.serve(app.fetch)
