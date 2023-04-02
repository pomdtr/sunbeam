export type Page = List | Detail;

export type List = {
  type: "list";
  showDetail?: boolean;
  items?: ListItem[];
};

export type Detail = {
  type: "detail";
  language?: string;
  text: string;
};

export type ListItem = {
  title: string;
  id?: string;
  subtitle?: string;
  actions?: Action[];
  accessories?: string[];
};

export type Action = OpenAction | RunAction | ReadAction | CopyAction;

export type ActionProps = {
  title: string;
  inputs?: Input[];
};

export type OpenAction = ActionProps & {
  type: "open";
  url?: string;
  path?: string;
};

export type RunAction = ActionProps & {
  type: "run";
  command: string;
  onSuccess?: "push" | "reload" | "exit";
};

export type CopyAction = ActionProps & {
  type: "copy";
  text: string;
};

export type ReadAction = ActionProps & {
  type: "read";
};

export type Input = TextField | TextArea | DropDown;

export type InputProps = {
  title: string;
  name: string;
};

export type TextField = InputProps & {
  type: "textfield";
  placeholder?: string;
  default?: string;
};

export type TextArea = InputProps & {
  type: "textarea";
  placeholder?: string;
  default?: string;
};

export type DropDown = InputProps & {
  type: "dropdown";
  choices: string[];
};
