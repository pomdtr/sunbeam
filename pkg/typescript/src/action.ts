import { Input } from "./input";

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
    mac?: string;
    windows?: string;
    linux?: string;
  };
  exit?: boolean;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  params?: Record<string, string | number | boolean | Input>;
  reload?: boolean;
  exit?: boolean;
} & ActionProps;

export type EditAction = {
  type: "edit";
  command: string;
  params?: Record<string, string | number | boolean | Input>;
} & ActionProps;

export type ReloadAction = {
  type: "reload";
  command: string;
  params?: Record<string, string | number | boolean | Input>;
} & ActionProps;


export type ExitAction = {
  type: "exit";
} & ActionProps;

export type Action = CopyAction | OpenAction | RunAction | ExitAction | EditAction | ReloadAction;
