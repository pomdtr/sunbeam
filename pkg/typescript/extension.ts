import { Commandspec, Manifest } from "./manifest.ts";
import { Command, Page } from "./page.ts";

export type CommandInput = {
  params: Record<string, string | number | boolean>;
  query?: string;
  formData?: Record<string, string | number | boolean>;
};

type CommandOutput = Promise<Page | Command | void> | Page | Command | void;

type RunFn = (
  input: CommandInput,
) => CommandOutput;

export class Extension {
  manifest: Manifest;
  private runners: Record<string, RunFn> = {};

  constructor(props: Omit<Manifest, "commands">) {
    this.manifest = {
      ...props,
      commands: [],
    };
  }

  static fromManifest(
    manifest: Manifest,
    run: (
      command: string,
      input: CommandInput,
    ) => CommandOutput,
  ) {
    const extension = new Extension(manifest);
    for (const command of manifest.commands) {
      extension.addCommand(command, (input) =>
        run(
          command.name,
          input || {
            params: {},
          },
        ));
    }
  }

  addCommand(
    props: Commandspec,
    run: (
      input: CommandInput,
    ) => CommandOutput,
  ) {
    this.manifest.commands.push(props);
    this.runners[props.name] = run;
    return this;
  }

  command(name: string) {
    return this.manifest.commands.find((c) => c.name === name);
  }

  run(name: string, input: CommandInput) {
    const run = this.runners[name];
    if (!run) {
      throw new Error(`Command ${name} not found`);
    }
    return run(
      input || {
        params: {},
      },
    );
  }

  toJSON() {
    return this.manifest;
  }
}
