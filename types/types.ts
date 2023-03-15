export type Page = List | Detail;

export type List = {
  type: "list";
  title?: string;
  items: ListItem[];
  actions: Action[];
};

export type Detail = {
  type: "detail";
  title?: string;
  actions: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions: Action[];
};

export type Action = RunAction;

type ActionProps = {
  title?: string;
  shortcut?: string;
};

export type RunAction = ActionProps & {
  type: "run";
  command: string;
  onSuccess?: "reload" | "replace" | "push";
};

export type OpenAction = ActionProps & {
  type: "open";
} & ({ url: string } | { path: string });

export type ReadAction = ActionProps & {
  type: "read";
  path: string;
};

export type CopyAction = ActionProps & {
  type: "copy";
  text: string;
};
