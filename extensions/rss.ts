#!/usr/bin/env deno run -A

import Parser from "npm:rss-parser";
import { formatDistance } from "npm:date-fns";
import * as sunbeam from "npm:sunbeam-types@0.25.1"
import { NodeHtmlMarkdown } from "npm:node-html-markdown"

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "RSS",
        description: "Manage your RSS feeds",
        items: [
            { command: "list" }
        ],
        requirements: [
            {
                name: "deno",
                link: "https://deno.com"
            }
        ],
        commands: [
            {
                name: "list",
                title: "Search Feeds",
                mode: "list",
            },
            {
                name: "show",
                title: "Show a feed",
                mode: "list",
                params: [
                    {
                        name: "url",
                        title: "URL",
                        required: true,
                        type: "text",
                    },
                ],
            },
            {
                name: "read",
                title: "Read a feed article",
                mode: "detail",
                hidden: true,
                params: [
                    {
                        name: "html",
                        title: "HTML",
                        required: true,
                        type: "textarea",
                    }
                ]
            }
        ]
    };

    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
if (payload.command == "list") {
    const feeds = payload.preferences.feeds as Record<string, string> || {};

    const list: sunbeam.List = {
        emptyText: "No feeds",
        items: Object.entries(feeds).map(([title, url]) => ({
            title,
            subtitle: url,
            actions: [
                {
                    title: "Show",
                    type: "run",
                    command: "show",
                    params: {
                        url
                    },
                },
                {
                    title: "Copy URL",
                    type: "copy",
                    key: "c",
                    text: url,
                    exit: true
                }
            ]
        })),
    }

    console.log(JSON.stringify(list));
} else if (payload.command == "show") {
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
                    title: "Read",
                    type: "run",
                    command: "read",
                    params: {
                        html: item.content || item.contentSnippet || ""
                    },
                },
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
    }

    console.log(JSON.stringify(page));
} else if (payload.command == "read") {
    const html = payload.params.html as string;

    const markdown = NodeHtmlMarkdown.translate(html)

    const detail = {
        text: markdown,
        format: "markdown",
        actions: [
            {
                title: "Copy",
                type: "copy",
                key: "c",
                text: markdown,
                exit: true
            }
        ]
    } as sunbeam.Detail;

    console.log(JSON.stringify(detail));
}
