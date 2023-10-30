#!/usr/bin/env -S deno run -A
import type * as sunbeam from "npm:sunbeam-types@0.23.15";

if (Deno.args.length == 0) {
    const manifest: sunbeam.Manifest = {
        title: "Tailscale",
        description: "Manage your tailscale devices",
        commands: [
            {
                name: "list-devices",
                title: "Search My Devices",
                mode: "list",
            },
            {
                name: "ssh-to-device",
                title: "SSH to Device",
                mode: "tty",
                params: [
                    {
                        name: "ip",
                        required: true,
                        description: "Device IP",
                        type: "string",
                    }
                ]
            }
        ],
    };
    console.log(JSON.stringify(manifest));

    Deno.exit(0);
}

type Device = {
    TailscaleIPs: string[];
    DNSName: string;
    OS: string;
    Online: boolean;
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.CommandInput;

if (payload.command == "list-devices") {
    const command = new Deno.Command("tailscale", { args: ["status", "--json"] });
    const { stdout } = await command.output()
    const status = JSON.parse(new TextDecoder().decode(stdout));
    const devices: Device[] = Object.values(status.Peer);
    const items: sunbeam.ListItem[] = devices.map((device) => ({
        title: device.DNSName.split(".")[0],
        subtitle: device.TailscaleIPs[0],
        accessories: [device.OS, device.Online ? "online" : "offline"],
        actions: [
            {
                title: "SSH to Device",
                type: "run",
                command: "ssh-to-device",
                params: {
                    ip: device.TailscaleIPs[0]
                },
            },
            {
                title: "Copy SSH Command",
                type: "copy",
                text: `ssh ${device.TailscaleIPs[0]}`,
                exit: true,
            },
            {
                title: "Copy IP",
                type: "copy",
                text: device.TailscaleIPs[0],
                key: "i",
                exit: true,
            },
        ],
    }));

    const list: sunbeam.List = { items };

    console.log(JSON.stringify(list));
} else if (payload.command == "ssh-to-device") {
    const params = payload.params as { ip: string };
    const command = new Deno.Command("ssh", { args: [params.ip] })
    const ps = command.spawn()
    await ps.status
}
