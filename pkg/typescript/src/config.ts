import type { RootItem } from "./manifest.ts"
export type Config = {
    $schema?: string;
    oneliners?: Oneliner[];
    extensions?: Record<string, ExtensionConfig>;
}

export type Oneliner = {
    title: string;
    command: string;
    exit?: boolean;
    dir?: string;
}

export type ExtensionConfig = {
    origin: string;
    preferences?: Record<string, string | number | boolean>;
    items?: RootItem[];
}

