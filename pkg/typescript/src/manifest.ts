export type Manifest = {
  title: string;
  description: string;
  preferences?: readonly Input[];
  commands: readonly Command[];
};

export type Command = {
  name: string;
  hidden?: boolean;
  title: string;
  params?: readonly Input[];
  mode: "filter" | "search" | "detail" | "tty" | "silent";
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

type PreferenceName<M extends Manifest> = NonNullable<
  M["preferences"]
>[number]["name"];

type PreferenceByName<
  M extends Manifest,
  N extends PreferenceName<M>,
> = Extract<NonNullable<M["preferences"]>[number], { name: N }>;

type ParamByName<
  M extends Manifest,
  N extends CommandName<M>,
  K extends ParamName<M, N>,
> = Extract<NonNullable<CommandByName<M, N>["params"]>[number], { name: K }>;

export type Payload<M extends Manifest> = {
  [N in CommandName<M>]:
    & {
      command: N;
      cwd: string;
      preferences: {
        [K in PreferenceName<M>]: PreferenceByName<M, K>["optional"] extends
          true ? InputMap[PreferenceByName<M, K>["type"]] | undefined
          : InputMap[PreferenceByName<M, K>["type"]];
      };
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
