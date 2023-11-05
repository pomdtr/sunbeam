import type { Action } from "./action";

export type List = {
  items: ListItem[];
  actions?: Action[];
  emptyText?: string;
  dynamic?: boolean;
};

export type Detail = {
  text: string;
  format?: "markdown" | "ansi" | "template";
  actions?: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions?: Action[];
};
