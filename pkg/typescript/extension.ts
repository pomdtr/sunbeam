import type { Command, Item, Manifest } from "./manifest.ts";
import type { Action, Page } from "./page.ts";

type RunProps = {
  params: Record<string, string | boolean>;
  query?: string;
};

type CommandProps = Command & {
  run: Runner;
  items?: Item[];
};

export type Runner = (
  props: RunProps,
) => Page | Action | void | Promise<Page | Action | void>;

export class Extension {
  private manifest: Omit<Manifest, "commands" | "items">;
  private _commands: Record<string, Command> = {};
  private _items: Item[] = [];
  private runners: Record<
    string,
    Runner
  > = {};

  constructor(props: Omit<Manifest, "commands" | "items">) {
    this.manifest = {
      ...props,
    };
  }

  static load(
    manifest: Manifest,
    runners: Record<string, Runner>,
  ) {
    const extension = new Extension({
      title: manifest.title,
      description: manifest.description,
      homepage: manifest.homepage,
    });
    extension._items.push(...manifest.items);

    for (const [name, command] of Object.entries(manifest.commands)) {
      const runner = runners[name];
      if (!runner) {
        throw new Error(`Command not found: ${name}`);
      }
      extension.command(name, {
        ...command,
        run: runner,
      });
    }

    return extension;
  }

  toJSON() {
    return {
      ...this.manifest,
      commands: Object.values(this._commands),
      items: this._items,
    };
  }

  get title() {
    return this.manifest.title;
  }

  get commands() {
    return this._commands;
  }

  get items() {
    return this._items;
  }

  get homepage() {
    return this.manifest.homepage;
  }

  command(
    name: string,
    props: CommandProps,
  ) {
    const { run, ...command } = props;
    this._commands[name] = command;
    this.runners[name] = run;
    if (props.items) {
      this._items.push(...props.items);
    }

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
