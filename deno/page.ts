import type { Action } from "./action.ts";

export type List = {
  items?: ListItem[];
  actions?: Action[];
  emptyText?: string;
  dynamic?: boolean;
};

export type Detail = {
  text: string;
  markdown?: boolean;
  actions?: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions?: Action[];
};
