export type Manifest = {
  title: string;
  commands: CommandSpec[];
};

export type CommandSpec = {
  name: string;
  title: string;
  mode: "view" | "no-view";
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
};

export type CommandInput = {
  params: Record<string, string | number | boolean>;
  query?: string;
};
