import { Action } from "./action.ts";

export type Config = {
  extensions: Record<string, {
    origin: string;
    env: Record<string, string>;
  }>;
  env: Record<string, string>;
  root: Action[];
};

export type ExtensionConfig = {
  origin: string;
  env: Record<string, string>;
};
