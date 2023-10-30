#!/usr/bin/env -S deno run -A

import { DB } from "https://deno.land/x/sqlite@v3.8/mod.ts";
import * as fs from "https://deno.land/std/fs/mod.ts";
import type * as sunbeam from "npm:sunbeam-types@0.23.7";
import * as path from "https://deno.land/std/path/mod.ts";

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "VS Code",
        description: "Manage your VS Code projects",
        commands: [
            {
                name: "list-projects",
                title: "List Projects",
                mode: "page",
            },
        ],
    };
    console.log(JSON.stringify(manifest));
    Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.CommandInput;

if (payload.command == "list-projects") {
    const homedir = Deno.env.get("HOME");
    const db = new DB(
        `${homedir}/Library/Application Support/Code/User/globalStorage/state.vscdb`,
    );
    const res = db.query(
        "SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'",
    );

    // deno-lint-ignore no-explicit-any
    let entries: any[] = JSON.parse(res[0][0] as string);
    entries = entries.filter((entry) => {
        if (!entry.folderUri) {
            return false;
        }

        const path = entry.folderUri.replace("file://", "");
        if (!fs.existsSync(path)) {
            return false;
        }

        return true;
    });


    const items: sunbeam.ListItem[] = entries.map((entry) => {
        const folderUri = new URL(entry.folderUri);
        const folderPath = path.dirname(folderUri.pathname);

        return {
            title: path.basename(folderUri.pathname),
            subtitle: path.basename(folderPath),
            actions: [
                {
                    title: "Open in VS Code",
                    type: "open",
                    target: entry.folderUri,
                    app: {
                        mac: "Visual Studio Code",
                    },
                    exit: true,
                },
                {
                    title: "Open Folder",
                    key: "o",
                    type: "open",
                    target: entry.folderUri,
                    exit: true
                },
                {
                    title: "Copy Path",
                    key: "c",
                    type: "copy",
                    exit: true,
                    text: entry.folderUri.replace("file://", ""),
                },
            ],
        }
    });

    const list: sunbeam.List = { type: "list", items };

    console.log(
        JSON.stringify(list),
    );
}
