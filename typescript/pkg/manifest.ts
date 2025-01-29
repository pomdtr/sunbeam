export type Manifest = {
  title: string;
  env?: string[];
  description?: string;
  commands: readonly Command[];
  root?: string;
};

export type Command = {
  name: string;
  hidden?: boolean;
  description?: string;
  params?: readonly Input[];
  mode: "filter" | "search" | "detail" | "silent";
};

export type Input = {
  name: string;
  title: string;
  type: "string" | "number" | "boolean";
  optional?: boolean;
};

type InputMap = {
  string: string;
  number: number;
  boolean: boolean;
};

type CommandName<M extends Manifest> = M["commands"][number]["name"];

type CommandByName<M extends Manifest, N extends CommandName<M>> = Extract<
  M["commands"][number],
  { name: N }
>;

type ParamName<M extends Manifest, N extends CommandName<M>> = NonNullable<
  CommandByName<M, N>["params"]
>[number]["name"];

type ParamByName<
  M extends Manifest,
  N extends CommandName<M>,
  K extends ParamName<M, N>,
> = Extract<NonNullable<CommandByName<M, N>["params"]>[number], { name: K }>;

export type Payload<M extends Manifest> = {
  [N in CommandName<M>]:
  & {
    command: N;
    params: CommandByName<M, N>["params"] extends undefined
    ? Record<string, never>
    : {
      [K in ParamName<M, N>]: ParamByName<M, N, K>["optional"] extends true
      ? InputMap[ParamByName<M, N, K>["type"]] | undefined
      : InputMap[ParamByName<M, N, K>["type"]];
    };
  }
  & (CommandByName<M, N>["mode"] extends "search" ? { query: string }
    : Record<string, never>);
}[CommandName<M>];
