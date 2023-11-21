#!/usr/bin/env -S deno run -A
import * as sunbeam from "https://raw.githubusercontent.com/pomdtr/sunbeam/main/sdk/mod.ts"

class ValTownClient {
    constructor(private token: string) { }

    fetch(url: string, init?: RequestInit) {
        return fetch(url, {
            ...init,
            headers: {
                ...init?.headers,
                "Authorization": `Bearer ${this.token}`
            }
        })
    }

    async fetchJSON(endpoint: string, init?: RequestInit) {
        const resp = await this.fetch(`https://api.val.town${endpoint}`, init);
        if (!resp.ok) {
            throw new Error("Failed to fetch");
        }

        return resp.json();
    }

    async pageinate(endpoint: string, init?: RequestInit) {
        let url = `https://api.val.town${endpoint}?limit=100`;
        let items: any = []
        while (true) {
            const resp = await client.fetch(url, init);
            if (!resp.ok) {
                throw new Error(`Failed to fetch: ${resp.statusText}`);
            }

            const { data, links } = await resp.json();
            items.push(...data);
            if (!links.next) {
                break;
            }
            url = links.next;
        }

        return items;
    }
}


if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Val Town",
        description: "Manage your Vals",
        preferences: [
            {
                name: "token",
                title: "Access Token",
                type: "text",
                required: true
            }
        ],
        root: ["list"],
        commands: [
            {
                title: "List Vals",
                name: "list",
                mode: "list",
                params: [
                    { name: "user", title: "User", required: false, type: "text" }
                ]
            },
            {
                title: "View Readme",
                name: "readme",
                mode: "detail",
                params: [
                    { name: "id", title: "Val ID", required: true, type: "text" }
                ]
            },
            {
                title: "View Source",
                name: "source",
                mode: "detail",
                params: [
                    { name: "id", title: "Val ID", required: true, type: "text" }
                ]
            }
        ]
    }

    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
const token = payload.preferences.token;
const client = new ValTownClient(token);

if (payload.command == "list") {
    const username = payload.params.user;
    const { id: userID } = await client.fetchJSON(username ? `/v1/alias/${username}` : "/v1/me")

    const vals = await client.pageinate(`/v1/users/${userID}/vals`);
    const items = vals.map((val: any) => {
        return {
            title: val.name,
            subtitle: `v${val.version}`,
            accessories: [
                val.privacy,
            ],
            actions: [
                {
                    "title": "Open in browser",
                    "type": "open",
                    "target": `https://val.town/v/${val.author.username.slice(1)}/${val.name}`
                },
                {
                    "title": "View readme",
                    "type": "run",
                    "command": "readme",
                    "params": {
                        "id": val.id
                    }
                },
                {
                    "title": "View source",
                    "type": "run",
                    "command": "source",
                    "params": {
                        "id": val.id
                    }
                }
            ]
        }
    })

    const list: sunbeam.List = {
        items
    }

    console.log(JSON.stringify(list));
} else if (payload.command == "readme") {
    const { readme } = await client.fetchJSON(`/v1/vals/${payload.params.id}`);
    const detail: sunbeam.Detail = {
        markdown: readme
    }
    console.log(JSON.stringify(detail));
} else if (payload.command == "source") {
    const { code } = await client.fetchJSON(`/v1/vals/${payload.params.id}`);
    const detail: sunbeam.Detail = {
        markdown: "```tsx\n" + code + "\n```"
    }
    console.log(JSON.stringify(detail));
}
