import type { Action } from "./action.ts";

export type List = {
  items?: ListItem[];
  actions?: Action[];
  showDetail?: boolean;
  emptyText?: string;
};

export type Detail = {
  text?: string;
  markdown?: string;
  actions?: Action[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  detail?: { text: string } | { markdown: string };
  actions?: Action[];
};
