#!/usr/bin/env -S deno run -A
import * as sunbeam from "https://raw.githubusercontent.com/pomdtr/sunbeam/main/sdk/mod.ts"

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Val Town",
        description: "Search and view Val Town vals",
        preferences: [
            {
                name: "token",
                title: "Access Token",
                type: "text",
                required: true
            }
        ],
        root: ["list", "search"],
        commands: [
            {
                title: "List Vals",
                name: "list",
                mode: "filter",
                params: [
                    { name: "user", title: "User", required: false, type: "text" }
                ]
            },
            {
                title: "Search Vals",
                name: "search",
                mode: "search",
            },
            {
                title: "View Readme",
                name: "readme",
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

async function run(payload: sunbeam.Payload) {
    const token = payload.preferences.token;
    const client = new ValTownClient(token);
    if (payload.command == "list") {
        const username = payload.params.user;
        const { id: userID } = await client.fetchJSON(username ? `/v1/alias/${username}` : "/v1/me")

        const vals = await client.paginate(`/v1/users/${userID}/vals`);
        const items = vals.map(valToListItem)

        const list: sunbeam.List = {
            showDetail: true,
            items
        }

        console.log(JSON.stringify(list));
    } else if (payload.command == "search") {
        const query = payload.query;
        if (query) {
            const { data: vals } = await client.fetchJSON(`/v1/search/vals?query=${encodeURIComponent(query)}&limit=50`);
            console.log(JSON.stringify({ showDetail: true, items: vals.map(valToListItem), emptyText: "No results" }));
        } else {
            console.log(JSON.stringify({ emptyText: "No query" }));
        }
    } else if (payload.command == "readme") {
        const { readme } = await client.fetchJSON(`/v1/vals/${payload.params.id}`);
        const detail: sunbeam.Detail = {
            markdown: readme || "No readme",
            actions: readme ? [
                { type: "copy", title: "Copy Readme", text: readme, exit: true }
            ] : []
        }
        console.log(JSON.stringify(detail));
    }
}

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

    async paginate(endpoint: string, init?: RequestInit) {
        const url = new URL(`https://api.val.town${endpoint}`)
        url.searchParams.set("limit", "100");

        let link: string = url.toString();
        const items: any = []
        while (true) {
            const resp = await this.fetch(link, init);
            if (!resp.ok) {
                throw new Error(`Failed to fetch: ${resp.statusText}`);
            }

            const { data, links } = await resp.json();
            items.push(...data);
            if (!links.next) {
                break;
            }
            link = links.next;
        }

        return items;
    }
}

function valToListItem(val: any): sunbeam.ListItem {
    return {
        title: val.name,
        subtitle: val.author.username,
        detail: {
            markdown: "```tsx\n" + val.code + "\n```"
        },
        actions: [
            {
            title: "Open in Browser",
            type: "open",
                url: `https://val.town/v/${val.author.username.slice(1)}/${val.name}`
            },
            {
                title: "Open Web Endpoint",
                type: "open",
                url: `https://${val.author.username.slice(1)}-${val.name}.web.val.run`
            },
            {
                title: "Copy URL",
                type: "copy",
                text: `https://val.town/v/${val.author.username.slice(1)}/${val.name}`
            },
            {
                title: "Copy Web Endpoint",
                type: "copy",
                text: `https://${val.author.username.slice(1)}-${val.name}.web.val.run`
            },
            {
                "title": "View Readme",
                "type": "run",
                "command": "readme",
                "key": "s",
                "params": {
                    "id": val.id
                }
            }
        ]
    }
}

try {
    await run(JSON.parse(Deno.args[0]));
} catch (e) {
    console.error(e);
    Deno.exit(1);
}
