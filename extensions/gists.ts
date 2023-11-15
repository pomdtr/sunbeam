#!/usr/bin/env -S deno run -A

import type * as sunbeam from "npm:sunbeam-types@0.25.1"
import * as path from "https://deno.land/std/path/mod.ts";

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Gists",
        items: [
            { command: "search" },
            { command: "create" }
        ],
        commands: [
            {
                name: "search",
                title: "Search Gists",
                mode: "list",
            },
            {
                name: "create",
                title: "Create Gist",
                mode: "tty",
                params: [
                    { name: "filename", title: "Filename", required: true, placeholder: "gist.md", type: "text" },
                    { name: "description", title: "Description", required: false, placeholder: "Gist Description", type: "text" },
                    { name: "public", title: "Public", required: false, type: "checkbox", label: "Whether the gist is public or not." },
                ]
            },
            {
                name: "browse",
                title: "Browse Gist Files",
                mode: "list",
                params: [
                    { name: "id", title: "Gist ID", required: true, type: "text" }
                ]
            },
            {
                name: "view",
                title: "View Gist File",
                mode: "detail",
                params: [
                    { name: "id", title: "Gist ID", required: true, type: "text" },
                    { name: "filename", title: "Filename", required: true, type: "text" }
                ]
            },
            {
                name: "edit",
                title: "Edit Gist File",
                mode: "tty",
            },
            {
                name: "delete",
                title: "Delete Gist",
                mode: "silent",
                params: [
                    { name: "id", title: "Gist ID", required: true, type: "text" }
                ]
            }
        ]
    }

    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}


const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
const githubToken = payload.preferences?.token as string;
if (!githubToken) {
    console.error("No github token set");
    Deno.exit(1);
}

try {
    const res = await run(payload);
    if (res) {
        console.log(JSON.stringify(res));
    }
} catch (e) {
    console.error(e);
    Deno.exit(1);
}


async function run(payload: sunbeam.Payload) {
    switch (payload.command) {
        case "search": {
            const resp = await fetchGithub("/gists");
            if (resp.status != 200) {
                throw new Error("Failed to fetch gists");
            }

            const gists = await resp.json() as any[];
            return {
                items: gists.map((gist) => ({
                    title: Object.keys(gist.files)[0],
                    subtitle: gist.description || "",
                    accessories: [
                        gist.public ? "Public" : "Private",
                    ],
                    actions: [
                        Object.keys(gist.files).length > 1 ? {
                            type: "run",
                            title: "Browse Files",
                            command: "browse",
                            params: {
                                id: gist.id
                            }
                        } : {
                            type: "run",
                            title: "View File",
                            command: "view",
                            params: {
                                id: gist.id,
                                filename: Object.keys(gist.files)[0]
                            }
                        },
                        {
                            type: "open",
                            title: "Open in Browser",
                            target: gist.html_url,
                            exit: true
                        },
                        {
                            type: "copy",
                            title: "Copy URL",
                            key: "c",
                            text: gist.html_url,
                            exit: true
                        },
                        {
                            title: "Create Gist",
                            key: "n",
                            type: "run",
                            command: "create",
                        },
                        {
                            title: "Delete Gist",
                            key: "d",
                            type: "run",
                            command: "delete",
                            reload: true,
                            params: {
                                id: gist.id
                            }
                        }
                    ]
                }))
            } as sunbeam.List;
        }
        case "create": {
            const filename = payload.params.filename as string;
            const content = await editor(payload.params.filename as string);
            const resp = await fetchGithub("/gists", {
                method: "POST",
                body: JSON.stringify({
                    description: payload.params.description,
                    public: payload.params.public,
                    files: {
                        [filename]: {
                            content
                        }
                    }
                })
            })

            if (resp.status != 201) {
                throw new Error("Failed to create gist");
            }
            return;
        }
        case "browse": {
            const id = payload.params.id as string;
            const resp = await fetchGithub(`/gists/${id}`);
            if (resp.status != 200) {
                throw new Error("Failed to fetch gist");
            }

            const gist = await resp.json() as any;
            return {
                items: Object.entries(gist.files).map(([filename, file]) => ({
                    title: filename,
                    actions: [
                        {
                            title: "View",
                            type: "run",
                            command: "view",
                            params: {
                                id,
                                filename
                            }
                        },
                        {
                            title: "Edit",
                            type: "run",
                            command: "edit",
                            params: {
                                id,
                                filename
                            }
                        }
                    ]
                }))
            } as sunbeam.List;
        }
        case "view": {
            const id = payload.params.id as string;
            const filename = payload.params.filename as string;
            const resp = await fetchGithub(`/gists/${id}`);
            if (resp.status != 200) {
                throw new Error("Failed to fetch gist");
            }

            const gist = await resp.json() as any;
            const file = gist.files[filename];
            if (!file) {
                throw new Error("File not found");
            }
            const lang = file.language?.toLowerCase();

            return {
                format: "markdown",
                text: lang == "md" ? file.content : `# ${filename}\n\n\`\`\`${lang || ""}\n${file.content}\n\`\`\``,
                actions: [
                    {
                        "title": "Copy Content",
                        "key": "c",
                        "type": "copy",
                        "text": file.content,
                        exit: true
                    },
                    {
                        title: "Copy Raw URL",
                        key: "r",
                        type: "copy",
                        text: file.raw_url,
                        exit: true
                    },
                    {
                        title: "Open in Browser",
                        type: "open",
                        target: gist.html_url,
                        exit: true
                    },
                ]
            } as sunbeam.Detail;
        }
        case "delete": {
            const id = payload.params.id as string;
            const resp = await fetchGithub(`/gists/${id}`, {
                method: "DELETE"
            });
            if (resp.status != 204) {
                throw new Error("Failed to delete gist");
            }
        }

    }
}


async function editor(filename: string) {
    const extension = path.extname(filename);
    const command = new Deno.Command("sunbeam", {
        args: ["edit", "--extension", extension],
        stdin: "null",
        stderr: "inherit",
        stdout: "piped",
    })

    const { success, stdout } = await command.output();
    if (!success) {
        throw new Error("Editor failed");
    }

    return new TextDecoder().decode(stdout);
}

function fetchGithub(endpoint: string, init?: RequestInit) {
    return fetch(`https://api.github.com${endpoint}`, {
        ...init,
        headers: {
            ...init?.headers,
            "Authorization": `Bearer ${githubToken}`
        }
    })

}
