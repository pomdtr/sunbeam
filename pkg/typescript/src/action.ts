export type Action = {
  title?: string;
  extension?: string;
  reload?: boolean;
  command: string;
  params?: Record<string, string | boolean | number | undefined>;
};
