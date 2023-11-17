import type { RootItem } from "./manifest.ts"
export type Config = {
    oneliners: Record<string, string>;
    extensions: Record<string, ExtensionConfig>;
}
export type ExtensionConfig = {
    origin: string;
    items?: RootItem[];
    install?: string;
    update?: string;

}

