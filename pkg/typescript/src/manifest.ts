export type Manifest = {
  title: string;
  commands: Command[];
};

type Command = {
  name: string;
  title: string;
  mode: "view" | "no-view" | "tty";
  hidden?: boolean;
  description?: string;
  params?: CommandParam[];
};

type CommandParam = StringParam | NumberParam | BooleanParam;

type ParamsProps = {
  name: string;
  description?: string;
  required?: boolean;
};

type StringParam = {
  type: "string";
  default?: string;
} & ParamsProps;

type NumberParam = {
  type: "number";
  default?: number;
} & ParamsProps;

type BooleanParam = {
  type: "boolean";
  default?: boolean;
} & ParamsProps;
