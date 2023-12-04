#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts"
import * as base64 from "https://deno.land/std@0.202.0/encoding/base64.ts";


if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "GitHub",
        description: "Search GitHub repositories",
        root: ["search"],
        preferences: [
            {
                name: "token",
                title: "Personal Access Token",
                type: "text",
                required: true
            }
        ],
        commands: [
            {
                title: "Search Repositories",
                name: "search",
                mode: "search"
            },
            {
                title: "List Issues",
                name: "issue.list",
                mode: "filter",
                params: [
                    {
                        name: "repo",
                        title: "The repository to list issues for",
                        type: "text",
                        required: true
                    }
                ]
            },
            {
                title: "List Pull Requests",
                name: "pr.list",
                mode: "filter",
                params: [
                    {
                        name: "repo",
                        title: "The repository to list pull requests for",
                        type: "text",
                        required: true
                    }
                ]
            },
            {
                title: "View Readme",
                name: "readme",
                mode: "detail",
                params: [
                    {
                        name: "repo",
                        title: "The repository to view the readme for",
                        type: "text",
                        required: true
                    }
                ]
            }
        ]
    }

    console.log(JSON.stringify(manifest, null, 2));
    Deno.exit(0);
}

const payload: sunbeam.Payload = JSON.parse(Deno.args[0]);
if (payload.preferences.token == null) {
    console.error("Missing required preference: token");
    Deno.exit(1);
}

try {
    await run(payload);
} catch (err) {
    console.error(err);
    Deno.exit(1);
}

async function run(payload: sunbeam.Payload) {
    const token = payload.preferences.token as string;
    if (payload.command == "search") {
        const query = payload.query
        if (!query) {
            const list: sunbeam.List = {
                emptyText: "Enter a query to search for repositories",
                items: []
            }
            console.log(JSON.stringify(list, null, 2));
            return;
        }

        const resp = await fetch(`https://api.github.com/search/repositories?q=${encodeURIComponent(query)}`, {
            headers: {
                "Authorization": `token ${token}`
            }
        });

        if (!resp.ok) {
            throw new Error(`Failed to search repositories: ${resp.status} ${resp.statusText}`);
        }

        const data = await resp.json();
        const list: sunbeam.List = {
            items: data.items.map((item: any) => ({
                title: item.full_name,
                accessories: [
                    `${item.stargazers_count} *`,
                ],
                actions: [
                    {
                        title: "View README",
                        type: "run",
                        command: "readme",
                        params: {
                            repo: item.full_name
                        }
                    },
                    {
                        title: "Open In Browser",
                        key: "o",
                        type: "open",
                        url: item.html_url,
                        exit: true
                    },
                    {
                        title: "List Issues",
                        key: "i",
                        type: "run",
                        command: "issue.list",
                        params: {
                            repo: item.full_name
                        }
                    },
                    {
                        title: "List Pull Requests",
                        key: "p",
                        type: "run",
                        command: "pr.list",
                        params: {
                            repo: item.full_name
                        }
                    },
                    {
                        title: "Copy URL",
                        key: "c",
                        type: "copy",
                        text: item.html_url,
                        exit: true
                    }
                ]
            } as sunbeam.ListItem))
        }

        console.log(JSON.stringify(list, null, 2));
    } else if (payload.command == "issue.list") {
        const repo = payload.params.repo;
        if (!repo) {
            throw new Error("Missing required parameter: repo");
        }

        const resp = await fetch(`https://api.github.com/repos/${repo}/issues`, {
            headers: {
                "Authorization": `token ${token}`
            }
        });

        if (!resp.ok) {
            throw new Error(`Failed to list issues: ${resp.status} ${resp.statusText}`);
        }

        const data = await resp.json();
        const list: sunbeam.List = {
            items: data.map((item: any) => ({
                title: item.title,
                accessories: [
                    `#${item.number}`,
                ],
                actions: [
                    {
                        title: "Open In Browser",
                        type: "open",
                        url: item.html_url,
                        exit: true
                    },
                    {
                        title: "Copy URL",
                        key: "c",
                        type: "copy",
                        text: item.html_url,
                        exit: true
                    }
                ]
            } as sunbeam.ListItem))
        }

        console.log(JSON.stringify(list, null, 2));
    } else if (payload.command == "pr.list") {
        const repo = payload.params.repo as string;
        const resp = await fetch(`https://api.github.com/repos/${repo}/pulls`, {
            headers: {
                "Authorization": `token ${token}`
            }
        });

        if (!resp.ok) {
            throw new Error(`Failed to list pull requests: ${resp.status} ${resp.statusText}`);
        }

        const data = await resp.json();
        const list: sunbeam.List = {
            items: data.map((item: any) => ({
                title: item.title,
                accessories: [
                    `#${item.number}`,
                ],
                actions: [
                    {
                        title: "Open In Browser",
                        type: "open",
                        url: item.html_url,
                        exit: true
                    },
                    {
                        title: "Copy URL",
                        key: "c",
                        type: "copy",
                        text: item.html_url,
                        exit: true
                    }
                ]
            } as sunbeam.ListItem))
        }

        console.log(JSON.stringify(list, null, 2));
    } else if (payload.command == "readme") {
        const repo = payload.params.repo as string;
        const resp = await fetch(`https://api.github.com/repos/${repo}/readme`, {
            headers: {
                "Authorization": `token ${token}`
            }
        });

        if (!resp.ok) {
            throw new Error(`Failed to view readme: ${resp.status} ${resp.statusText}`);
        }

        const data = await resp.json();
        const markdown = new TextDecoder().decode(base64.decode(data.content));

        const detail: sunbeam.Detail = {
            markdown,
            actions: [
                {
                    title: "Open in Browser",
                    type: "open",
                    url: data.html_url,
                    exit: true
                },
                {
                    title: "List Issues",
                    key: "i",
                    type: "run",
                    command: "issue.list",
                    params: {
                        repo: repo
                    }
                },
                {
                    title: "List Pull Requests",
                    key: "p",
                    type: "run",
                    command: "pr.list",
                    params: {
                        repo: repo
                    }
                }
            ]
        }

        console.log(JSON.stringify(detail, null, 2));
    } else {
        throw new Error(`Unknown command: ${payload.command}`);
    }
}

