import type { Input } from "./input";
export type Manifest = {
  title: string;
  description?: string;
  root?: CommandRef[];
  commands: CommandSpec[];
};

type CommandRef = {
  title: string;
  command: string;
  params?: Record<string, string | number | boolean | Input>;
};

export type CommandSpec = {
  name: string;
  title: string;
  mode: "list" | "detail" | "tty" | "silent";
  hidden?: boolean;
  description?: string;
  params?: CommandParam[];
};

type CommandParam = StringParam | BooleanParam | NumberParam;

type ParamsProps = {
  name: string;
  description?: string;
  required?: boolean;
};

type StringParam = {
  type: "string";
  default?: string;
} & ParamsProps;

type BooleanParam = {
  type: "boolean";
  default?: boolean;
} & ParamsProps;

type NumberParam = {
  type: "number";
  default?: number;
} & ParamsProps;

type InputParams = Record<string, string | number | boolean>;
export type CommandInput<T extends InputParams = InputParams> = {
  command: string;
  params: T;
  query?: string;
  cwd?: string;
};
