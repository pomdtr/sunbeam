import { Input } from "./input";

export type Manifest = {
  title: string;
  platforms?: Platform[];
  description?: string;
  requirements?: Requirement[];
  preferences?: Input[];
  root?: CommandRef[];
  commands: CommandSpec[];
};

type Platform = "linux" | "macos" | "windows";

type Requirement = {
  name: string;
  link?: string;
};

type CommandRef = {
  command: string;
  title?: string;
  params?: Record<string, string | number | boolean>;
};

export type CommandSpec = {
  name: string;
  title: string;
  mode: "list" | "detail" | "tty" | "silent";
  hidden?: boolean;
  description?: string;
  params?: Input[];
};

type PayloadParams = Record<string, string | boolean>;
export type Payload<T extends PayloadParams = PayloadParams, V extends PayloadParams = PayloadParams> = {
  command: string;
  params: T;
  preferences: V;
  query?: string;
  cwd: string;
};
