type ActionProps = {
  title: string;
  key?: string;
}

export type CopyAction = {
  type: "copy";
  text: string;
  exit?: boolean;
} & ActionProps;

export type OpenAction = {
  type: "open";
  target: string;
  app?: {
    macos?: string;
    windows?: string;
    linux?: string;
  };
  exit?: boolean;
} & ActionProps;

export type EditAction = {
  type: "edit";
  target: string;
  exit?: boolean;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  params?: Record<string, string | number | boolean>;
  reload?: boolean;
  exit?: boolean;
} & ActionProps;


export type ReloadAction = {
  type: "reload";
  command: string;
  params?: Record<string, string | number | boolean>;
} & ActionProps;


export type ExitAction = {
  type: "exit";
} & ActionProps;

export type Action = CopyAction | OpenAction | RunAction | ExitAction | EditAction | ReloadAction;
