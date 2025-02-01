type ActionProps = {
  title: string;
};

export type CopyAction = {
  type: "copy";
  text: string;
} & ActionProps;

export type OpenAction = {
  type: "open";
  url: string
} & ActionProps;

export type EditAction = {
  type: "edit";
  path: string;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  params?: Params;
  reload?: boolean;
} & ActionProps;

export type ReloadAction = {
  type: "reload";
  params?: Params;
} & ActionProps;

export type Params = Record<string, string | number | boolean>;

export type Action =
  | CopyAction
  | OpenAction
  | RunAction
  | ReloadAction;
