#!/usr/bin/env -S deno run -A

import type * as sunbeam from "npm:sunbeam-sdk@0.2.1"
import * as dates from "npm:date-fns"

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Deno Deploy",
        description: "Manage your Deno Deploy projects",
        items: [
            { command: "projects" }
        ],
        preferences: [
            {
                name: "token",
                title: "Access Token",
                type: "text",
                required: true
            }
        ],
        commands: [
            {
                name: "projects",
                title: "List Projects",
                mode: "list",
            },
            {
                name: "deployments",
                title: "List Deployments",
                mode: "list",
                params: [
                    { name: "project", title: "Project", required: true, type: "text" }
                ]
            },
            {
                name: "playground",
                title: "View Playground",
                mode: "detail",
                params: [
                    { name: "project", title: "Project", required: true, type: "text" }
                ]
            }
        ]
    }

    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
const deployToken = payload.preferences.token;
if (!deployToken) {
    console.error("Missing deploy token");
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
        case "projects": {
            const resp = await fetchDeployAPI("/projects");
            if (resp.status != 200) {
                throw new Error("Failed to fetch projects");
            }
            const projects = await resp.json();

            return {
                items: projects.map((project: any) => {
                    const item: sunbeam.ListItem = {
                        title: project.name,
                        accessories: [project.type],
                        actions: []
                    }

                    if (project.type == "git") {
                        const repo = project.git.repository
                        item.actions?.push({
                            title: "List Deployments",
                            type: "run",
                            command: "deployments",
                            params: {
                                project: project.name,
                            }
                        }, {
                            title: "Open Repository",
                            type: "open",
                            target: `https://github.com/${repo.owner}/${repo.name}`,
                            exit: true,
                        })
                    } else if (project.type == "playground") {
                        item.actions?.push({
                            title: "View Playground",
                            type: "run",
                            command: "view-playground",
                            params: {
                                project: project.name,
                            }
                        }, {
                            title: "Open Playground",
                            type: "open",
                            target: `https://dash.deno.com/playground/${project.id}`,
                            exit: true,
                        })
                    }

                    if (project.hasProductionDeployment) {
                        const domains = project.productionDeployment.deployment.domainMappings
                        const domain = domains.length ? domains[domains.length - 1].domain : "No domain"
                        item.subtitle = domain
                        item.actions?.push({
                            title: "Open Production URL",
                            type: "open",
                            target: `https://${domain}`,
                            exit: true,
                        })
                    }

                    item.actions?.push({
                        title: "Open Dashboard",
                        type: "open",
                        target: `https://dash.deno.com/projects/${project.id}`,
                        exit: true,
                    }, {
                        title: "Copy Dashboard URL",
                        type: "copy",
                        key: "c",
                        text: `https://dash.deno.com/projects/${project.id}`,
                        exit: true,
                    })


                    return item
                })
            } as sunbeam.List;
        }
        case "playground": {
            const name = payload.params.project as string;
            const resp = await fetchDeployAPI(`/projects/${name}`);
            if (resp.status != 200) {
                throw new Error("Failed to fetch project");
            }

            const project = await resp.json();
            if (project.type != "playground") {
                throw new Error("Project is not a playground");
            }

            const snippet = project.playground.snippet;
            const lang = project.playground.mediaType
            return {
                format: "markdown",
                text: `\`\`\`${lang}\n${snippet}\n\`\`\``,
                actions: [
                    {
                        title: "Copy Snippet",
                        key: "c",
                        type: "copy",
                        text: snippet,
                        exit: true
                    },
                    {
                        title: "Open in Browser",
                        type: "open",
                        target: `https://dash.deno.com/playground/${project.id}`,
                        exit: true
                    }
                ],
            } as sunbeam.Detail;
        }
        case "deployments": {
            const project = payload.params.project as string;

            const resp = await fetchDeployAPI(`/projects/${project}/deployments`);
            if (resp.status != 200) {
                throw new Error("Failed to fetch deployments");
            }

            const [deployments] = await resp.json();
            return {
                items: deployments.map(({ id, createdAt, deployment, relatedCommit }: any) => {
                    const item = {
                        title: id,
                        accessories: [dates.formatDistance(new Date(createdAt), new Date(), {
                            addSuffix: true,
                        })],
                        actions: [],
                    } as sunbeam.ListItem;

                    if (deployment.domainMappings?.length) {
                        item.actions?.push({
                            title: "Open URL",
                            type: "open",
                            target: `https://${deployment.domainMappings[0].domain}`,
                            exit: true,
                        })
                    }

                    if (relatedCommit) {
                        item.title = relatedCommit.message;
                        item.actions?.push({
                            title: "Open Commit",
                            type: "open",
                            target: relatedCommit.url,
                            exit: true,
                        })
                    }

                    return item;
                })
            } as sunbeam.List;
        }
        case "logs": {
            const { project, deployment } = payload.params as { project: string, deployment: string };
            const resp = await fetchDeployAPI(`/projects/${project}/deployments/${deployment}`);
            if (resp.status != 200) {
                throw new Error("Failed to fetch deployment");
            }


        }
    }
}

function fetchDeployAPI(endpoint: string, init?: RequestInit) {
    return fetch(`https://dash.deno.com/api${endpoint}`, {
        ...init,
        headers: {
            ...init?.headers,
            "Authorization": `Bearer ${deployToken}`,
        }
    })
}
