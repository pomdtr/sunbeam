import type { Action } from "./action";

export type List = {
  title?: string;
  items: ListItem[];
  actions?: Action[];
  emptyText?: string;
  dynamic?: boolean;
};

export type Detail = {
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
