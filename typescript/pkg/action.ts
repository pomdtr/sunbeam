type ActionProps = {
  title: string;
};

export type CopyAction = {
  type: "copy";
  text: string;
} & ActionProps;

export type OpenAction = {
  type: "open";
  target: string
} & ActionProps;

export type EditAction = {
  type: "edit";
  path: string;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  params?: Record<string, Param>;
  reload?: boolean;
} & ActionProps;

export type Param =
  | string
  | number
  | boolean
  | { default?: string | number | boolean; optional?: boolean };

export type ReloadAction = {
  type: "reload";
  params?: Record<string, Param>;
} & ActionProps;

export type Action =
  | CopyAction
  | OpenAction
  | RunAction
  | EditAction
  | ReloadAction;
