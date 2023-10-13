import type { Command } from "./command";
export type Page = List | Detail | Form;

export type List = {
  type: "list";
  title?: string;
  items: ListItem[];
  actions?: Action[];
  emptyText?: string;
  reload?: boolean;
};

export type Detail = {
  type: "detail";
  title?: string;
  markdown: string;
  actions?: Action[];
};

export type Form = {
  type: "form";
  title?: string;
  fields: FormField[];
};

export type ListItem = {
  title: string;
  subtitle?: string;
  accessories?: string[];
  actions: Action[];
};

export type FormField = {
  title: string;
  name: string;
  required?: boolean;
  input: FormInput;
};

type FormInput = TextField | TextArea | Checkbox | Select;

type TextField = {
  type: "text";
  placeholder?: string;
  default?: string;
  secure?: boolean;
};

type TextArea = {
  type: "textarea";
  placeholder?: string;
  default?: string;
};

type Checkbox = {
  type: "checkbox";
  label: string;
  default?: boolean;
};

type Select = {
  type: "select";
  options: SelectOption[];
  default?: string;
};

type SelectOption = {
  title: string;
  value: string;
};

export type Action = {
  title: string;
  key?: string;
  onAction: Command;
};
