import type { Command, Manifest } from "./manifest.ts";

type RunProps = {
  params: Record<string, string | boolean>;
  query?: string;
};

export class Extension {
  private manifest: Omit<Manifest, "commands">;
  private _commands: Record<string, Command> = {};
  private runners: Record<string, (props: RunProps) => unknown> = {};

  constructor(props: Omit<Manifest, "commands">) {
    this.manifest = {
      ...props,
    };
  }

  toJSON() {
    return { ...this.manifest, commands: Object.values(this._commands) };
  }

  get title() {
    return this.manifest.title;
  }

  get commands() {
    return this._commands;
  }

  get homepage() {
    return this.manifest.homepage;
  }

  command(
    params: Command & {
      run: (props: RunProps) => unknown;
    },
  ) {
    const { run, ...command } = params;

    this._commands[command.name] = command;
    this.runners[command.name] = run;

    return this;
  }

  run(name: string, props?: RunProps) {
    const runner = this.runners[name];
    if (!runner) {
      throw new Error(`Command not found: ${name}`);
    }

    return runner(props || { params: {} });
  }
}

export function createExtension(props: Omit<Manifest, "commands">) {
  return new Extension(props);
}

export function loadExtension(
  manifest: Manifest,
  run: (command: string, props: RunProps) => unknown,
) {
  const extension = new Extension({
    title: manifest.title,
    description: manifest.description,
  });

  for (const command of manifest.commands) {
    extension.command({
      ...command,
      run: (props) => run(command.name, props),
    });
  }

  return extension;
}
