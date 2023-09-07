import type { Manifest } from "./manifest.ts";

export class Extension {
  static async fetch(origin: string) {
    const resp = await fetch(origin);
    if (!resp.ok) {
      throw new Error(`Failed to fetch manifest from ${origin}`);
    }

    const manifest = await resp.json();
    return new Extension(manifest as Manifest);
  }

  constructor(public manifest: Manifest) {}

  get title() {
    return this.manifest.title;
  }

  get commands() {
    return this.manifest.commands;
  }

  get entrypoint() {
    return this.manifest.entrypoint;
  }

  command(name: string) {
    return this.commands.find((c) => c.name === name);
  }

  async run(
    name: string,
    options?: { args?: Record<string, string | boolean>; query?: string }
  ) {
    const command = this.command(name);
    if (!command) {
      throw new Error(`Command ${name} not found`);
    }

    const entrypoint = this.entrypoint;
    if (!entrypoint) {
      throw new Error(`Entrypoint not found`);
    }

    const resp = await fetch(entrypoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        command: name,
        ...options,
      }),
    });

    if (!resp.ok) {
      throw new Error(`Failed to run command ${name}`);
    }

    return resp.json();
  }
}
