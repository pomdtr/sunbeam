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

type InputProps = {
  name: string;
  required: boolean;
}

type Textfield = InputProps & {
  type: "text";
  title: string;
  defaut?: string;
  placeholder?: string;
}

type TextArea = InputProps & {
  type: "textarea";
  title: string;
  defaut?: string;
  placeholder?: string;
}

type Password = InputProps & {
  type: "password";
  title: string;
  defaut?: string;
  placeholder?: string;
}

type Checkbox = InputProps & {
  type: "checkbox";
  label: string;
  title?: string;
  defaut?: boolean;
}

export type Input = Textfield | TextArea | Password | Checkbox;
