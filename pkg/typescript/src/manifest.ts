export type Manifest = {
  title: string;
  description?: string;
  commands: CommandSpec[];
};

export type CommandSpec = {
  name: string;
  title: string;
  mode: "page" | "silent" | "tty";
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
  params: T;
  query?: string;
};
