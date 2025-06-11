import type { Params } from "./action.ts";

export function parseArgs(args: string[]): { command: string; params: Params } {
  if (args.length === 0) {
    return { command: "", params: {} };
  }

  const command = args[0];
  const params: Params = {};

  for (let i = 1; i < args.length; i += 2) {
    const key = args[i];
    const value = args[i + 1];

    if (!key.startsWith("--")) {
      continue;
    }

    const paramName = key.slice(2);
    
    // Try to parse as boolean first, then number, otherwise keep as string
    if (value === "true") {
      params[paramName] = true;
    } else if (value === "false") {
      params[paramName] = false;
    } else if (value !== "" && !isNaN(Number(value)) && !isNaN(parseFloat(value))) {
      params[paramName] = Number(value);
    } else {
      params[paramName] = value;
    }
  }

  return { command, params };
}