import type { Command, Manifest } from "./manifest.ts";
import { readAll } from "https://deno.land/std@0.201.0/streams/read_all.ts";

type RunProps = {
  params?: Record<string, string | boolean>;
  query?: string;
};

export class Extension {
  private manifest: Manifest;
  private runners: Record<string, (props: RunProps) => unknown> = {};

  constructor(props: Omit<Manifest, "commands">) {
    this.manifest = {
      ...props,
      commands: [],
    };
  }

  toJSON() {
    return this.manifest;
  }

  get title() {
    return this.manifest.title;
  }

  addCommand(
    params: Command & {
      run: (props: RunProps) => unknown;
    }
  ) {
    const { run, ...command } = params;

    this.manifest.commands.push(command);
    this.runners[command.name] = run;

    return this;
  }

  get commands() {
    return this.manifest.commands;
  }

  command(name: string) {
    return this.commands.find((c) => c.name === name);
  }

  run(name: string, props?: RunProps) {
    const runner = this.runners[name];
    if (!runner) {
      throw new Error(`Command not found: ${name}`);
    }

    return runner(props || {});
  }

  // here we need an arrow function to preserve the `this` context when using Deno.serve(extension.fetch)
  fetch = async (req: Request) => {
    if (req.method === "GET") {
      return Response.json(this.manifest);
    }

    if (req.method !== "POST") {
      return Response.json(
        {
          error: "Invalid request",
        },
        { status: 400 }
      );
    }

    const url = new URL(req.url);
    const [, command] = url.pathname.split("/");
    if (!command) {
      return Response.json(
        {
          error: "Invalid request: missing command",
        },
        {
          status: 400,
        }
      );
    }

    const body = await req.json();
    const props = body as {
      params?: Record<string, string | boolean>;
      query?: string;
    };
    if (!command) {
      return Response.json(
        {
          error: "Invalid request",
        },
        {
          status: 400,
        }
      );
    }

    const runner = this.runners[command];
    if (!runner) {
      return Response.json(
        {
          error: `Command not found: ${command}`,
        },
        {
          status: 404,
        }
      );
    }

    try {
      const result = await runner(props);
      return Response.json(result);
    } catch (e) {
      return Response.json(
        {
          error: e.message,
        },
        {
          status: 500,
        }
      );
    }
  };

  async execute(args: string[]) {
    if (args.length === 0) {
      console.log(JSON.stringify(this.manifest));
      return;
    }

    const command = this.command(args[0]);
    if (!command) {
      throw new Error(`Command not found: ${args[0]}`);
    }

    const stdin = new TextDecoder().decode(await readAll(Deno.stdin));
    const params = JSON.parse(stdin);

    if (!params || typeof params !== "object") {
      throw new Error("Invalid params");
    }

    for (const param of command.params || []) {
      if (param.optional) {
        continue;
      }

      if (!(param.name in params)) {
        throw new Error(`Missing required param: ${param.name}`);
      }
    }

    try {
      const res = await this.run(command.name, params);
      if (command.mode !== "silent") {
        console.log(JSON.stringify(res));
      }
    } catch (e) {
      throw new Error(`Failed to execute command: ${e.message}`);
    }
  }
}

export function createExtension(props: Omit<Manifest, "commands">) {
  return new Extension(props);
}

export async function fetchExtension(origin: string) {
  const resp = await fetch(origin);
  if (!resp.ok) {
    throw new Error(`Failed to fetch extension manifest from ${origin}`);
  }

  const manifest = (await resp.json()) as Manifest;
  return loadExtension(origin, manifest);
}

export function loadExtension(origin: string, manifest: Manifest) {
  const extension = new Extension({
    title: manifest.title,
    description: manifest.description,
  });

  for (const command of manifest.commands) {
    extension.addCommand({
      ...command,
      run: async (props?: RunProps) => {
        const url = new URL(command.name, origin);
        const resp = await fetch(url, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(props || {}),
        });
        if (!resp.ok) {
          throw new Error(
            `Failed to fetch ${url}: ${resp.status} ${resp.statusText}`
          );
        }
        return resp.json();
      },
    });
  }

  return extension;
}
