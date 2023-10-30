import type { Action } from "./action";
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
  text: string;
  highlight?: "markdown" | "ansi";
  actions?: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions: Action[];
};
