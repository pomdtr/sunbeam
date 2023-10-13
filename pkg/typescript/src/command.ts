export type Command = CopyCommand | OpenCommand | RunCommand | ExitCommand;

export type CopyCommand = {
  type: "copy";
  text: string;
  exit?: boolean;
};

export type OpenCommand = {
  type: "open";
  target: string;
  exit?: boolean;
};

export type RunCommand = {
  type: "run";
  command: string;
  exit?: boolean;
};

export type ExitCommand = {
  type: "exit";
};
