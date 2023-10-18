import type { Command } from "./command";
export type Page = List | Detail;

export type List = {
  type: "list";
  title?: string;
  items: ListItem[];
  actions?: Action[];
  emptyText?: string;
  dynamic?: boolean;
};

export type Detail = {
  type: "detail";
  title?: string;
  markdown: string;
  actions?: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions: Action[];
};

export type Action = {
  title: string;
  key?: string;
  onAction: Command;
};
