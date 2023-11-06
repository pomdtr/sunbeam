#!/usr/bin/env deno run -A

import { Octokit } from "https://esm.sh/octokit@3.1.1?dts";
import * as sunbeam from "npm:sunbeam-types@0.23.21"

if (Deno.args.length === 0) {
    const manifest: sunbeam.Manifest = {
        title: "Gist",
        description: "Manage your gists",
        requirements: [
            {
                name: "deno",
                link: "https://deno.com"
            }
        ],
        env: [
            {
                name: "GITHUB_TOKEN",
                description: "GitHub Personal Access Token",
                required: true,
            }
        ],
        commands: [
            {
                name: "list",
                title: "List Gists",
                mode: "list",
            },
            {
                name: "browse",
                title: "Browser Gist Files",
                mode: "list",
                params: [
                    {
                        name: "id",
                        description: "Gist ID",
                        type: "string",
                        required: true,
                    }
                ]
            },
            {
                name: "delete",
                title: "Delete Gist",
                mode: "silent",
                params: [
                    {
                        name: "id", description: "Gist ID", type: "string", required: true,
                    }
                ]
            },
            {
                name: "view",
                title: "View Gist File",
                mode: "detail",
                params: [
                    {
                        name: "id",
                        description: "Gist ID",
                        type: "string",
                        required: true,
                    },
                    {
                        name: "file",
                        description: "File Name",
                        type: "string",
                        required: true,
                    }
                ]
            }
        ]
    }
    console.log(JSON.stringify(manifest))
    Deno.exit(0)
}

const oktokit = new Octokit({ auth: Deno.env.get("GITHUB_TOKEN") });
const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload
if (payload.command == "list") {
    const gists = await oktokit.request("GET /gists");
    const items = gists.data.map((gist) => {
        const files = Object.values(gist.files)
        return {
            title: files.length > 0 ? Object.values(gist.files)[0].filename! : "Untitled",
            subtitle: gist.description || "",
            actions: [
                files.length > 1 ? {
                    title: "Browse Files",
                    type: "run",
                    command: "browse",
                    params: {
                        id: gist.id,
                    }
                } : {
                    title: "View File",
                    type: "run",
                    command: "view",
                    params: {
                        id: gist.id,
                        file: Object.values(gist.files)[0].filename!
                    }
                },
                {
                    title: "Delete Gist",
                    type: "run",
                    key: "d",
                    command: "delete",
                    reload: true,
                    params: {
                        id: gist.id,
                    }
                },
                {
                    title: "Open in Browser",
                    type: "open",
                    target: gist.html_url,
                    exit: true,
                },
                {
                    title: "Copy ID",
                    type: "copy",
                    text: gist.id,
                    exit: true,
                }
            ]
        } as sunbeam.ListItem
    })

    console.log(JSON.stringify({ items }))
    Deno.exit(0)
} else if (payload.command == "browse") {
    const params = payload.params as {
        id: string,
    }
    const gist = await oktokit.request("GET /gists/{gist_id}", {
        gist_id: params.id,
    });

    const files = Object.values(gist.data.files || {})
    const items: sunbeam.ListItem[] = files.map((file) => {
        return {
            title: file!.filename || "Untitled",
            subtitle: file!.language || "",
            actions: [
                {
                    title: "View File",
                    type: "run",
                    command: "view",
                    params: {
                        id: gist.data.id || "",
                        file: file!.filename!
                    }
                },
                {
                    title: "Copy Raw URL",
                    type: "copy",
                    text: file!.raw_url || "",
                    exit: true,
                }
            ]
        }
    })

    console.log(JSON.stringify({ items }))
    Deno.exit(0)
} else if (payload.command == "delete") {
    const params = payload.params as {
        id: string,
    }

    await oktokit.request("DELETE /gists/{gist_id}", {
        gist_id: params.id,
    });

    Deno.exit(0)
} else if (payload.command == "view") {
    const params = payload.params as {
        id: string,
        file: string,
    }

    const gist = await oktokit.request("GET /gists/{gist_id}", {
        gist_id: params.id,
    });

    const file = gist.data.files![params.file]

    const page: sunbeam.Detail = {
        format: "markdown",
        text: ["```" + file?.language?.toLowerCase() || "txt", file?.content, "```"].join("\n"),
        actions: [
            {
                title: "Copy Raw URL",
                type: "copy",
                text: file!.raw_url || "",
                exit: true,
            },
            {
                title: "Copy Content",
                type: "copy",
                text: file!.content || "",
                exit: true,
            }
        ]
    }

    console.log(JSON.stringify(page))
    Deno.exit(0)
}

