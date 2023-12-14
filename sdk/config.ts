import { Param } from "./action.ts";

export type Config = {
  $schema?: string;
  oneliners?: Record<string, Oneliner>;
  extensions?: Record<string, ExtensionConfig>;
};

export type Oneliner = {
  command: string;
  exit?: boolean;
  cwd?: string;
};

export type ExtensionConfig = {
  origin: string;
  preferences?: Record<string, string | number | boolean>;
  root?: Record<string, RootItem[]>;
};

export type RootItem = {
  command: string;
  params?: Record<string, Param>;
};
