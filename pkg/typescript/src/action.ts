type ActionProps = {
  title?: string;
  key?: string;
};

export type CopyAction = {
  type: "copy";
  text: string;
  exit?: boolean;
} & ActionProps;

export type OpenAction = {
  type: "open";
  url?: string;
  path?: string;
} & ActionProps;

export type EditAction = {
  type: "edit";
  path: string;
  exit?: boolean;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  params?: Record<string, Param>;
  reload?: boolean;
  exit?: boolean;
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

export type ExitAction = {
  type: "exit";
} & ActionProps;

export type Action =
  | CopyAction
  | OpenAction
  | RunAction
  | ExitAction
  | EditAction
  | ReloadAction;
