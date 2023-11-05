#!/usr/bin/env deno run -A

import Parser from "npm:rss-parser";
import { formatDistance } from "npm:date-fns";
import * as sunbeam from "npm:sunbeam-types@0.23.15"

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "RSS",
        description: "Manage your RSS feeds",
        commands: [
            {
                name: "show",
                title: "Show a feed",
                mode: "list",
                params: [
                    {
                        name: "url",
                        description: "URL",
                        required: true,
                        type: "string",
                    },
                ],
            },
        ]
    };

    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.CommandInput;
if (payload.command == "show") {
    const params = payload.params as { url: string };
    const feed = await new Parser().parseURL(params.url);
    const page: sunbeam.List = {

        items: feed.items.map((item) => ({
            title: item.title || "",
            subtitle: item.categories?.join(", ") || "",
            accessories: item.isoDate
                ? [
                    formatDistance(new Date(item.isoDate), new Date(), {
                        addSuffix: true,
                    }),
                ]
                : [],
            actions: [
                {
                    title: "Open in browser",
                    type: "open",
                    target: item.link || "",
                    exit: true
                },
                {
                    title: "Copy Link",
                    type: "copy",
                    key: "c",
                    text: item.link || "",
                    exit: true
                },
            ],
        })),
    };

    console.log(JSON.stringify(page));
}
