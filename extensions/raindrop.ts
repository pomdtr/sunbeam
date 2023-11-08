#!/usr/bin/env -S deno run -A
import * as sunbeam from "npm:sunbeam-types@0.23.27"

if (Deno.args.length === 0) {
    const manifest: sunbeam.Manifest = {
        title: "Raindrop",
        description: "Manage your raindrop bookmarks",
        requirements: [
            {
                name: "deno",
                link: "https://deno.com"
            }
        ],
        preferences: [
            {
                name: "token",
                description: "Raindrop API token",
                type: "string",
                required: true,
            }
        ],
        commands: [
            {
                title: "Search Bookmarks",
                name: "search-bookmarks",
                mode: "list",
            },
        ],
    };
    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
const raindropToken = payload.preferences.token as string;

if (payload.command == "search-bookmarks") {
    const resp = await fetch("https://api.raindrop.io/rest/v1/raindrops/0", {
        headers: {
            Authorization: `Bearer ${raindropToken}`,
        },
    });

    const { items: bookmarks } = await resp.json() as {
        items: {
            title: string;
            link: string;
            domain: string;
        }[]
    }

    const list: sunbeam.List = {
        items: bookmarks.map((bookmark) => ({
            title: bookmark.title,
            subtitle: bookmark.domain,
            actions: [
                {
                    title: "Open URL",
                    type: "open",
                    target: bookmark.link,
                    exit: true,
                },
                {
                    title: "Copy URL",
                    key: "c",
                    type: "copy",
                    text: bookmark.link,
                    exit: true,
                },
            ],
        }))
    }

    console.log(JSON.stringify(list));
} else {
    console.error(`Unknown command: ${payload.command}`);
    Deno.exit(1);
}
