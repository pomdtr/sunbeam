#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Tabs",
        description: "Manage Browser Tabs",
        commands: [
            {
                "name": "list",
                "title": "List Tabs",
                mode: "filter"
            },
            {
                name: "close",
                title: "Close tab",
                mode: "silent",
                params: [
                    {
                        name: "id",
                        title: "Tab ID",
                        type: "number",
                        required: true
                    }
                ]
            },
            {
                name: "focus",
                title: "Focus tab",
                mode: "silent",
                params: [
                    {
                        name: "id",
                        title: "Tab ID",
                        type: "number",
                        required: true
                    }
                ]
            }
        ]
    }
    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}
const payload: sunbeam.Payload = JSON.parse(Deno.args[0]);

if (payload.command == "list") {
    const { stdout, success } = await new Deno.Command("popcorn", {
        args: ["tab", "list", "--json"]
    }).outputSync();

    if (!success) {
        console.error("Error: Failed to list tabs");
        Deno.exit(1);
    }

    const tabs = JSON.parse(new TextDecoder().decode(stdout));
    const list: sunbeam.List = {
        items: tabs.map((tab: any) => ({
            title: tab.title || "Untitled",
            subtitle: tab.url || "",
            actions: [
                {
                    type: "run",
                    title: "Focus Tab",
                    exit: true,
                    command: "focus",
                    params: {
                        id: tab.id
                    }
                },
                {
                    type: "copy",
                    title: "Copy URL",
                    text: tab.url,
                    exit: true
                },
                {
                    type: "run",
                    title: "Close Tab",
                    command: "close",
                    reload: true,
                    params: {
                        id: tab.id
                    }
                }
            ]
        }))
    }

    console.log(JSON.stringify(list));
} else if (payload.command == "close") {
    const tabID = payload.params["id"] as number;

    const { success } = await new Deno.Command("popcorn", {
        args: ["tab", "close", tabID.toString()]
    }).output();

    if (!success) {
        console.error("Error: Failed to close tab");
        Deno.exit(1);
    }
} else if (payload.command == "focus") {
    const tabID = payload.params["id"] as number;

    const { success } = await new Deno.Command("popcorn", {
        args: ["tab", "focus", tabID.toString()]
    }).output();

    if (!success) {
        console.error("Error: Failed to focus tab");
        Deno.exit(1);
    }
}



