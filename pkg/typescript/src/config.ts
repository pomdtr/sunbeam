import type { RootItem } from "./manifest.ts"
export type Config = {
    oneliners: Record<string, string | Oneliner>;
    extensions: Record<string, ExtensionConfig>;
}

export type Oneliner = {
    command: string;
    exit?: boolean;
    dir?: string;
}

export type ExtensionConfig = {
    origin: string;
    items?: RootItem[];
    install?: string;
    update?: string;
}

