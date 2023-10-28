import { Input } from "./input";

type ActionProps = {
  title: string;
  key?: string;
}

export type CopyCommand = {
  type: "copy";
  text: string;
  exit?: boolean;
} & ActionProps;

export type OpenCommand = {
  type: "open";
  target: string;
  app?: {
    mac?: string;
    windows?: string;
    linux?: string;
  };
  exit?: boolean;
} & ActionProps;

export type RunCommand = {
  type: "run";
  command: string;
  params?: Record<string, string | number | boolean | Input>;
} & ActionProps;

export type ExitCommand = {
  type: "exit";
} & ActionProps;

export type Action = CopyCommand | OpenCommand | RunCommand | ExitCommand;
